import { browser } from '$app/environment';
import { env } from '$env/dynamic/public';
import { Capacitor } from '@capacitor/core';

/** Backend address used when nothing else is configured (local dev). */
const DEFAULT_API_URL = 'http://localhost:8080';

/**
 * The backend the phone app runs itself (EmbeddedBackend.java, loopback only).
 * The native app uses it unless another server is saved in Settings.
 */
const EMBEDDED_API_URL = 'http://127.0.0.1:47831';

/** How long the app waits for its on-device backend to come up at launch. */
const EMBEDDED_STARTUP_TIMEOUT_MS = 30_000;

/** localStorage key holding the backend address picked in Settings. */
const SERVER_URL_STORAGE_KEY = 'veil:server-url';

/** Hosts that only mean "this machine", so their URLs are useless to other devices. */
const LOOPBACK_HOSTS = ['localhost', '127.0.0.1', '[::1]'];

/** Whether the client runs inside the native (Capacitor) app. */
export function isNativeApp(): boolean {
	return browser && Capacitor.isNativePlatform();
}

/** The backend address saved on this device, if any. */
export function storedServerUrl(): string | null {
	if (!browser) return null;
	return localStorage.getItem(SERVER_URL_STORAGE_KEY);
}

/** Saves the backend address for this device; takes effect on the next page load. */
export function saveServerUrl(url: string): void {
	localStorage.setItem(SERVER_URL_STORAGE_KEY, normalizeServerUrl(url));
}

/** Forgets the saved address, so the native app goes back to its on-device backend. */
export function clearServerUrl(): void {
	localStorage.removeItem(SERVER_URL_STORAGE_KEY);
}

/** Whether the client talks to the backend running on this phone. */
export function usesEmbeddedBackend(): boolean {
	return isNativeApp() && storedServerUrl() === null;
}

/** Trims whitespace and trailing slashes and defaults the scheme to http. */
export function normalizeServerUrl(url: string): string {
	let normalized = url.trim().replace(/\/+$/, '');
	if (!/^https?:\/\//.test(normalized)) {
		normalized = `http://${normalized}`;
	}
	return normalized;
}

/**
 * Base URL of the Go backend: the device's saved address, else the on-device
 * backend in the native app, else the build's PUBLIC_API_URL in the browser or
 * INTERNAL_API_URL during SSR.
 */
export function apiBase(): string {
	if (!browser) {
		const internal = process.env.INTERNAL_API_URL;
		if (internal) return internal;
		return DEFAULT_API_URL;
	}
	const stored = storedServerUrl();
	if (stored) return stored;
	if (isNativeApp()) return EMBEDDED_API_URL;
	if (env.PUBLIC_API_URL) return env.PUBLIC_API_URL;
	return DEFAULT_API_URL;
}

/**
 * Points a backend-built URL at the address this client reaches the backend on.
 * The backend stamps blob and stream URLs with its PUBLIC_URL, which defaults to
 * localhost; from another device (the phone app) that host is unreachable.
 */
export function backendUrl(url: string): string {
	if (!url.startsWith('http://') && !url.startsWith('https://')) return url;
	const parsed = new URL(url);
	if (!LOOPBACK_HOSTS.includes(parsed.hostname)) return url;
	if (!parsed.pathname.startsWith('/api/')) return url;
	const base = apiBase();
	if (parsed.origin === new URL(base).origin) return url;
	return `${base}${parsed.pathname}${parsed.search}`;
}

/** A plugin's icon (a cached blob URL, a remote URL or an emoji), reachable from this client. */
export function pluginIconUrl(iconUrl: string | null): string | null {
	if (iconUrl === null) return null;
	return backendUrl(iconUrl);
}

/** Whether the backend answers its health check at the given address. */
export async function checkServer(url: string): Promise<boolean> {
	try {
		const response = await fetch(`${normalizeServerUrl(url)}/health`, {
			signal: AbortSignal.timeout(5000)
		});
		return response.ok;
	} catch {
		return false;
	}
}

/**
 * Waits until the on-device backend answers, which takes a moment after the
 * app starts. Returns false if it doesn't come up in time.
 */
export async function waitForEmbeddedBackend(): Promise<boolean> {
	const deadline = Date.now() + EMBEDDED_STARTUP_TIMEOUT_MS;
	while (Date.now() < deadline) {
		if (await checkServer(EMBEDDED_API_URL)) return true;
		await new Promise((resolve) => setTimeout(resolve, 250));
	}
	return false;
}
