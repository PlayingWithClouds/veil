/**
 * Publishes plugin bundles to the GitHub repository's rolling "plugins"
 * release, where every backend picks them up within its next update check
 * (pluginstore: at start and every 6 h). Each version is one npm-style tarball
 * (package/package.json + package/plugin.js) uploaded as a release asset; the
 * release's index.json names every plugin's latest version and is uploaded
 * last, so backends never see a version whose tarball isn't there yet. Only
 * the bundle is published, not the TypeScript source: bump a plugin's
 * "version" in its package.json, then run this. Versions already in the index
 * are skipped. Needs the gh CLI, logged in with push access to the repository.
 *
 * Usage: bun scripts/publish-plugins.ts [--dry-run] [plugin ...]
 *   plugin     source folder names (default: all)
 *   --dry-run  pack and show what would be published, without uploading
 */
import { mkdtemp, rename, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { bundlePlugin, folderName, pluginNames, sourcePackage } from './lib/bundle';

const REPOSITORY = 'PlayingWithClouds/veil';
/** The release tag holding the plugin assets (pluginstore.DefaultIndex). */
const RELEASE_TAG = 'plugins';
const INDEX_FILE = 'index.json';

/** One index entry: a package's latest version and where its tarball lives. */
interface IndexEntry {
	name: string;
	version: string;
	description?: string;
	main: string;
	keywords: string[];
	dist: { tarball: string; integrity: string };
}

interface PluginIndex {
	packages: IndexEntry[];
}

/** Runs a command, returning its exit code and stdout. */
async function run(command: string[], cwd?: string): Promise<{ exitCode: number; stdout: string }> {
	const child = Bun.spawn(command, { cwd, stdout: 'pipe', stderr: 'inherit' });
	const stdout = await new Response(child.stdout).text();
	return { exitCode: await child.exited, stdout };
}

/** Runs a command and throws when it fails. */
async function runChecked(command: string[], cwd?: string): Promise<string> {
	const { exitCode, stdout } = await run(command, cwd);
	if (exitCode !== 0) throw new Error(`${command.join(' ')} exited with ${exitCode}`);
	return stdout;
}

/** Creates the plugins release unless it exists; it is never marked latest. */
async function ensureRelease(): Promise<void> {
	const { exitCode } = await run(['gh', 'release', 'view', RELEASE_TAG, '-R', REPOSITORY, '--json', 'tagName']);
	if (exitCode === 0) return;
	await runChecked([
		'gh', 'release', 'create', RELEASE_TAG, '-R', REPOSITORY,
		'--title', 'Plugins',
		'--notes', 'Plugin bundles installed and updated by Veil backends. index.json lists the latest version of each.',
		'--latest=false'
	]);
}

/** The published index, or an empty one before the first publish. */
async function fetchIndex(): Promise<PluginIndex> {
	const { exitCode, stdout } = await run(['gh', 'release', 'download', RELEASE_TAG, '-R', REPOSITORY, '-p', INDEX_FILE, '-O', '-']);
	if (exitCode !== 0 || stdout.trim() === '') {
		return { packages: [] };
	}
	return JSON.parse(stdout) as PluginIndex;
}

/** Packs a bundled plugin folder into an npm-style tarball at tarballPath. */
async function packTarball(bundleFolder: string, tarballPath: string): Promise<void> {
	const packRoot = await mkdtemp(join(tmpdir(), 'veil-pack-'));
	try {
		await rename(bundleFolder, join(packRoot, 'package'));
		await runChecked(['tar', '-czf', tarballPath, '-C', packRoot, 'package']);
	} finally {
		await rm(packRoot, { recursive: true, force: true });
	}
}

/** The npm-style "sha512-<base64>" integrity of a file. */
async function integrityOf(path: string): Promise<string> {
	const hasher = new Bun.CryptoHasher('sha512');
	hasher.update(await Bun.file(path).arrayBuffer());
	return 'sha512-' + hasher.digest('base64');
}

/** The download URL of a release asset. */
function assetUrl(assetName: string): string {
	return `https://github.com/${REPOSITORY}/releases/download/${RELEASE_TAG}/${assetName}`;
}

/** Replaces (or adds) a package's entry in the index. */
function upsert(index: PluginIndex, entry: IndexEntry): void {
	index.packages = index.packages.filter((existing) => existing.name !== entry.name);
	index.packages.push(entry);
	index.packages.sort((first, second) => first.name.localeCompare(second.name));
}

const dryRun = process.argv.includes('--dry-run');
const requested = process.argv.slice(2).filter((arg) => !arg.startsWith('--'));
let names = await pluginNames();
if (requested.length > 0) {
	names = names.filter((name) => requested.includes(name));
}

if (!dryRun) {
	await ensureRelease();
}
const index = await fetchIndex();
const stagingDir = await mkdtemp(join(tmpdir(), 'veil-plugins-'));
let changed = false;
try {
	for (const name of names) {
		const pkg = await sourcePackage(name);
		const published = index.packages.find((entry) => entry.name === pkg.name);
		if (published && published.version === pkg.version) {
			console.log(`${pkg.name}@${pkg.version} already published`);
			continue;
		}
		const assetName = `${folderName(pkg.name)}-${pkg.version}.tgz`;
		const tarballPath = join(stagingDir, assetName);
		const bundleFolder = await bundlePlugin(name, stagingDir, 'release');
		await packTarball(bundleFolder, tarballPath);

		if (!dryRun) {
			const { exitCode } = await run(['gh', 'release', 'upload', RELEASE_TAG, tarballPath, '-R', REPOSITORY, '--clobber']);
			if (exitCode !== 0) {
				console.error(`uploading ${pkg.name}@${pkg.version} failed`);
				process.exitCode = 1;
				continue;
			}
		}
		upsert(index, {
			name: pkg.name,
			version: pkg.version,
			description: pkg.description,
			main: 'plugin.js',
			keywords: ['veil-plugin'],
			dist: { tarball: assetUrl(assetName), integrity: await integrityOf(tarballPath) }
		});
		changed = true;
		console.log(`published ${pkg.name}@${pkg.version}${dryRun ? ' (dry run)' : ''}`);
	}

	if (changed && !dryRun) {
		const indexPath = join(stagingDir, INDEX_FILE);
		await Bun.write(indexPath, JSON.stringify(index, null, '\t') + '\n');
		await runChecked(['gh', 'release', 'upload', RELEASE_TAG, indexPath, '-R', REPOSITORY, '--clobber']);
		console.log(`updated ${INDEX_FILE} (${index.packages.length} plugins)`);
	}
} finally {
	await rm(stagingDir, { recursive: true, force: true });
}
