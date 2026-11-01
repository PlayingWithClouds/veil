import { test, expect } from '@playwright/test';

const BACKEND = process.env.PUBLIC_API_URL ?? 'http://localhost:8080';

/** Whether a backend answers, so the test can skip instead of failing without one. */
async function backendIsUp(): Promise<boolean> {
	try {
		const response = await fetch(`${BACKEND}/health`);
		return response.ok;
	} catch {
		return false;
	}
}

/** Stores a page of the first plugin that can list videos, so home has something. */
async function fillLibrary(): Promise<boolean> {
	const graphql = (query: string) =>
		fetch(`${BACKEND}/graphql`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ query })
		}).then((response) => response.json());
	const existing = await graphql('{ videos(limit: 1) { id } }');
	if (existing.data.videos.length > 0) return true;
	const plugins = await graphql('{ plugins { name enabled available capabilities } }');
	const lister = plugins.data.plugins.find(
		(plugin: { enabled: boolean; available: boolean; capabilities: string[] }) =>
			plugin.enabled && plugin.available && plugin.capabilities.includes('scene:list')
	);
	if (!lister) return false;
	await graphql(
		`mutation { browse(plugin: "${lister.name}", kind: VIDEO, limit: 5) { videos { id } } }`
	);
	return true;
}

test('home lists videos and a video page opens', async ({ page }) => {
	test.skip(!(await backendIsUp()), `no backend at ${BACKEND}`);
	test.skip(!(await fillLibrary()), 'no plugin can list videos');

	await page.goto('/');
	const firstCard = page.locator('main button.tv-card').first();
	await expect(firstCard).toBeVisible();
	await firstCard.click();

	await expect(page).toHaveURL(/\/video\//);
	await expect(page.locator('main h1')).not.toBeEmpty();
	await expect(page.getByText('Sources')).toBeVisible();
});
