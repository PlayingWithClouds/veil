import { gqlClient } from '$lib/veil';

const FALLBACK = {
	maxConcurrentJobs: 4,
	maxJobRetries: 3,
	downloadSpeedLimitKBps: 0,
	allowDownloadsWhileStreaming: true,
	autoEnrichAfterScrape: false,
	requireVpn: false,
	kindLimits: [] as Array<{
		kind: string;
		maxConcurrent: number;
		retryInitialMs: number;
		retryMultiplier: number;
		retryMaxMs: number;
	}>
};

export async function load() {
	const data = await gqlClient
		.query({
			settings: {
				maxConcurrentJobs: true,
				maxJobRetries: true,
				downloadSpeedLimitKBps: true,
				allowDownloadsWhileStreaming: true,
				autoEnrichAfterScrape: true,
				requireVpn: true,
				kindLimits: {
					kind: true,
					maxConcurrent: true,
					retryInitialMs: true,
					retryMultiplier: true,
					retryMaxMs: true
				}
			}
		})
		.catch(() => ({ settings: FALLBACK }));

	return { settings: data.settings };
}
