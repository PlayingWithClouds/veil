// When the backend says a site wants a login it couldn't renew (GraphQL
// error code NEEDS_LOGIN), wherever that happened, a toast links to the
// plugin's login on the plugins page.
import { showUndo } from '$lib/stores/undoToast';
import { pluginLoginUrl } from '$lib/routes';

/** How long a plugin isn't prompted for again, so a page of failing calls shows one toast. */
const PROMPT_PAUSE_MS = 30_000;

const promptedAt = new Map<string, number>();

/** One GraphQL error as the backend presents it. */
interface PresentedError {
	extensions?: { code?: string; plugin?: string };
}

/** fetch for the GraphQL client: answers as usual, prompting for a login an answer asks for. */
export async function fetchWithLoginPrompt(
	input: RequestInfo | URL,
	init?: RequestInit
): Promise<Response> {
	const response = await fetch(input, init);
	void promptForLogins(response.clone());
	return response;
}

/** Prompts for the logins a GraphQL answer says are needed. */
async function promptForLogins(response: Response) {
	const text = await response.text();
	if (!text.includes('NEEDS_LOGIN')) return;
	const errors = (JSON.parse(text) as { errors?: PresentedError[] }).errors ?? [];
	for (const error of errors) {
		const plugin = error.extensions?.plugin;
		if (error.extensions?.code === 'NEEDS_LOGIN' && plugin) promptLogin(plugin);
	}
}

/** Shows the toast for a plugin, unless it was just shown. */
function promptLogin(plugin: string) {
	const last = promptedAt.get(plugin) ?? 0;
	if (Date.now() - last < PROMPT_PAUSE_MS) return;
	promptedAt.set(plugin, Date.now());
	showUndo({
		message: `${plugin} needs a login`,
		link: { label: 'Log in', href: pluginLoginUrl(plugin) }
	});
}
