import { gqlClient } from '$lib/veil';
import { pluginIconUrl } from '$lib/server';

/** A loaded plugin as cards and pickers need it. */
export interface SitePlugin {
	name: string;
	displayName: string | null;
	iconUrl: string | null;
	capabilities: string[];
	enabled: boolean;
	available: boolean;
	// Hostnames the plugin serves, used to label cards with their site.
	domains: string[];
}

/** A plugin with everything the plugins page shows and edits. */
export interface InstalledPlugin extends SitePlugin {
	packageName: string | null;
	localBuild: boolean;
	description: string | null;
	version: string;
	content: string[];
	requiresSolver: boolean;
	settings: PluginSettingField[];
	settingValues: { key: string; value: string }[];
	// How the site is logged into; null for sites without accounts.
	login: PluginLogin | null;
	// This device's login on the site, if any.
	account: PluginAccount | null;
}

/** A site's login: a form, imported browser cookies, or both. */
export interface PluginLogin {
	fields: PluginSettingField[];
	cookies: boolean;
}

/** A login on a plugin's site, kept on this device. */
export interface PluginAccount {
	id: string;
	label: string;
	status: 'ok' | 'expired' | 'needsLogin' | 'failed';
	lastCheckedAt: string | null;
	// Logs in again by itself when the site ends the session.
	storesLogin: boolean;
}

/** One setting a plugin asks for. */
export interface PluginSettingField {
	key: string;
	label: string;
	description: string | null;
	type: string;
	required: boolean;
	default: string | null;
}

/** A plugin package in the published index. */
export interface PluginPackage {
	name: string;
	version: string;
	description: string | null;
	installedVersion: string | null;
}

/** What a plugin can list into the library. */
export type ListKind = 'VIDEO' | 'GALLERY' | 'PERFORMER' | 'STUDIO' | 'TAG' | 'AUDIO' | 'STORY';

/** A setting or login field as the plugins page shows it. */
const SETTING_FIELD = {
	key: true,
	label: true,
	description: true,
	type: true,
	required: true,
	default: true
} as const;

/** An account as the plugins page shows it. */
const ACCOUNT = {
	id: true,
	label: true,
	status: true,
	lastCheckedAt: true,
	storesLogin: true
} as const;

/** Every installed plugin with its settings and login. */
export async function fetchPlugins(): Promise<InstalledPlugin[]> {
	const data = await gqlClient.query({
		plugins: {
			name: true,
			displayName: true,
			iconUrl: true,
			capabilities: true,
			enabled: true,
			available: true,
			domains: true,
			packageName: true,
			localBuild: true,
			description: true,
			version: true,
			content: true,
			requiresSolver: true,
			settings: SETTING_FIELD,
			settingValues: { key: true, value: true },
			login: { fields: SETTING_FIELD, cookies: true },
			account: ACCOUNT
		}
	});
	return data.plugins.map((plugin) => ({
		...plugin,
		iconUrl: pluginIconUrl(plugin.iconUrl ?? null),
		login: plugin.login ?? null,
		account: plugin.account ?? null
	}));
}

/**
 * Logs into a plugin's site with its form's values. The backend keeps them
 * encrypted to log in again when the site ends the session.
 */
export async function loginAccount(
	plugin: string,
	values: Record<string, string>
): Promise<PluginAccount> {
	const data = await gqlClient.mutation({
		loginAccount: { __args: { plugin, values }, ...ACCOUNT }
	});
	return data.loginAccount;
}

/** Keeps cookies exported from a browser as a plugin's login; only its own site's are kept. */
export async function importAccountCookies(
	plugin: string,
	cookies: string
): Promise<PluginAccount> {
	const data = await gqlClient.mutation({
		importAccountCookies: { __args: { plugin, cookies }, ...ACCOUNT }
	});
	return data.importAccountCookies;
}

/** Asks the site whether a login still works. */
export async function checkAccount(id: string): Promise<PluginAccount> {
	const data = await gqlClient.mutation({ checkAccount: { __args: { id }, ...ACCOUNT } });
	return data.checkAccount;
}

/** Forgets a login, with its session and stored values. */
export async function logoutAccount(id: string): Promise<void> {
	await gqlClient.mutation({ logoutAccount: { __args: { id } } });
}

/** The plugins that can list this kind, enabled and available here. */
export function listers(plugins: SitePlugin[], kind: ListKind): SitePlugin[] {
	const capability = `${kind === 'VIDEO' ? 'scene' : kind.toLowerCase()}:list`;
	return plugins.filter(
		(plugin) => plugin.enabled && plugin.available && plugin.capabilities.includes(capability)
	);
}

/** The plugin serving a page URL, by its domains. */
export function pluginForUrl(url: string, plugins: SitePlugin[]): SitePlugin | null {
	let host: string;
	try {
		host = new URL(url).hostname.replace(/^www\./, '');
	} catch {
		return null;
	}
	for (const plugin of plugins) {
		if (plugin.domains.some((domain) => host === domain || host.endsWith(`.${domain}`)))
			return plugin;
	}
	return null;
}

/** Published packages matching query, all when empty. */
export async function fetchPluginPackages(query = ''): Promise<PluginPackage[]> {
	const data = await gqlClient.query({
		pluginPackages: {
			__args: { query },
			name: true,
			version: true,
			description: true,
			installedVersion: true
		}
	});
	return data.pluginPackages.map((found) => ({
		...found,
		description: found.description ?? null,
		installedVersion: found.installedVersion ?? null
	}));
}

/** Turns a plugin on or off. */
export async function setPluginEnabled(name: string, enabled: boolean): Promise<void> {
	await gqlClient.mutation({ setPluginEnabled: { __args: { name, enabled } } });
}

/** Saves a plugin's setting values. */
export async function savePluginSettings(
	plugin: string,
	values: { key: string; value: string }[]
): Promise<void> {
	await gqlClient.mutation({ updatePluginSettings: { __args: { plugin, values } } });
}

/** Installs (or updates) a published package. */
export async function installPlugin(packageName: string): Promise<void> {
	await gqlClient.mutation({ installPlugin: { __args: { packageName } } });
}

/** Removes an installed plugin. */
export async function uninstallPlugin(name: string): Promise<void> {
	await gqlClient.mutation({ uninstallPlugin: { __args: { name } } });
}

/** Installs newer versions of installed plugins; returns the updated package names. */
export async function updatePlugins(): Promise<string[]> {
	const data = await gqlClient.mutation({ updatePlugins: true });
	return data.updatePlugins;
}

/** A plugin's display name, else its name. */
export function siteDisplayName(pluginName: string, list: SitePlugin[]): string {
	const plugin = list.find((entry) => entry.name === pluginName);
	if (plugin?.displayName) return plugin.displayName;
	return pluginName;
}
