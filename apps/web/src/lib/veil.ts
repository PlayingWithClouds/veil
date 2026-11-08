import { browser } from '$app/environment';
import { createClient as createWsClient, type Client as WsClient } from 'graphql-ws/client';
// Import from the client subpath (not the barrel) so the SDK's Bun-specific
// plugin runtime never enters the browser build graph.
import { createVeilClient } from '@playingwithclouds/veil-sdk/client';
import { apiBase } from '$lib/server';
import { fetchWithLoginPrompt } from '$lib/loginPrompt';

/** GraphQL endpoint of the backend (public URL in the browser, internal in SSR). */
function resolveEndpoint(): string {
	return `${apiBase()}/graphql`;
}

// In the browser, answers saying a site wants a login link to it.
export const veil = createVeilClient({
	url: resolveEndpoint(),
	fetch: browser ? fetchWithLoginPrompt : undefined
});

/** Raw genql client for operations outside the CRUD facade. */
export const gqlClient = veil.client;

// GraphQL-over-WebSocket client for live subscriptions. Browser-only and lazily
// created so SSR never opens a socket. Uses the graphql-transport-ws subprotocol
// that gqlgen's websocket transport negotiates.
let wsClient: WsClient | null = null;

export function getWsClient(): WsClient | null {
	if (!browser) return null;
	if (wsClient) return wsClient;
	const httpUrl = resolveEndpoint();
	const wsUrl = httpUrl.replace(/^http/, 'ws');
	wsClient = createWsClient({ url: wsUrl, lazy: true, retryAttempts: Infinity, retryWait });
	return wsClient;
}

// How long the socket waits before connecting again after losing the backend.
const RETRY_WAIT_MS = 1000;

/**
 * Waits a second before each reconnect. The default doubles the wait, so a
 * search started after a backend restart could sit for 15 s or more.
 */
function retryWait(): Promise<void> {
	return new Promise((resolve) => setTimeout(resolve, RETRY_WAIT_MS));
}
