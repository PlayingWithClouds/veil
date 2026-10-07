import { gqlClient } from '$lib/veil';
import { pluginIconUrl } from '$lib/server';

/** Installed plugins, plus the plugin packages published on npm. */
export async function load() {
	const [installed, published] = await Promise.all([
		gqlClient
			.query({
				plugins: {
					id: true,
					name: true,
					packageName: true,
					localBuild: true,
					displayName: true,
					iconUrl: true,
					description: true,
					version: true,
					capabilities: true,
					enabled: true,
					requiresSolver: true,
					available: true,
					settings: {
						key: true,
						label: true,
						description: true,
						type: true,
						required: true,
						default: true
					},
					settingValues: { key: true, value: true }
				}
			})
			.catch(() => ({ plugins: [] })),
		// The registry is unreachable offline; the page still works without it.
		gqlClient
			.query({
				pluginPackages: { name: true, version: true, description: true, installedVersion: true }
			})
			.catch(() => ({ pluginPackages: [] }))
	]);

	const plugins = installed.plugins.map((plugin) => ({
		...plugin,
		iconUrl: pluginIconUrl(plugin.iconUrl)
	}));
	return { plugins, packages: published.pluginPackages };
}
