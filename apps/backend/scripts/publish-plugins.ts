/**
 * Publishes plugin bundles to npm, where every backend picks them up within
 * its next update check (pluginstore: at start and every 6 h). Only the
 * bundle is published, not the TypeScript source: bump a plugin's "version"
 * in its package.json, then run this. Versions already on npm are skipped.
 *
 * Usage: bun scripts/publish-plugins.ts [--dry-run] [plugin ...]
 *   plugin     source folder names (default: all)
 *   --dry-run  pack and show what would be published, without publishing
 */
import { mkdtemp, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { bundlePlugin, pluginNames, sourcePackage } from './lib/bundle';

/** Runs a command, returning its exit code and stdout. */
async function run(command: string[], cwd?: string): Promise<{ exitCode: number; stdout: string }> {
	const child = Bun.spawn(command, { cwd, stdout: 'pipe', stderr: 'inherit' });
	const stdout = await new Response(child.stdout).text();
	return { exitCode: await child.exited, stdout };
}

/** Whether name@version is already on the registry. */
async function isPublished(name: string, version: string): Promise<boolean> {
	const { exitCode, stdout } = await run(['npm', 'view', `${name}@${version}`, 'version']);
	return exitCode === 0 && stdout.trim() === version;
}

const dryRun = process.argv.includes('--dry-run');
const requested = process.argv.slice(2).filter((arg) => !arg.startsWith('--'));
let names = await pluginNames();
if (requested.length > 0) {
	names = names.filter((name) => requested.includes(name));
}

const stagingDir = await mkdtemp(join(tmpdir(), 'veil-plugins-'));
try {
	for (const name of names) {
		const pkg = await sourcePackage(name);
		if (await isPublished(pkg.name, pkg.version)) {
			console.log(`${pkg.name}@${pkg.version} already published`);
			continue;
		}
		const folder = await bundlePlugin(name, stagingDir, 'release');
		const publish = ['npm', 'publish', '--access', 'public'];
		if (dryRun) publish.push('--dry-run');
		const { exitCode } = await run(publish, folder);
		if (exitCode !== 0) {
			console.error(`publishing ${pkg.name}@${pkg.version} failed`);
			process.exitCode = 1;
			continue;
		}
		console.log(`published ${pkg.name}@${pkg.version}${dryRun ? ' (dry run)' : ''}`);
	}
} finally {
	await rm(stagingDir, { recursive: true, force: true });
}
