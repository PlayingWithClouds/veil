import { defineConfig } from '@playwright/test';

// Browser tests run against a dev server started here (a port of its own, so
// another project's vite on 5173 is never mistaken for it) and a backend you
// start yourself on :8080; without a backend they skip.
const PORT = 5174;

export default defineConfig({
	testDir: 'tests',
	timeout: 90_000,
	use: { baseURL: `http://localhost:${PORT}` },
	webServer: {
		command: `bun run dev --port ${PORT} --strictPort`,
		url: `http://localhost:${PORT}`,
		reuseExistingServer: true,
		timeout: 120_000
	}
});
