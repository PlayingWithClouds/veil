<!--
	A plugin's login on its site: whether this device is logged in, the
	site's login form (its values are kept encrypted, so the backend logs in
	again when the site ends the session), importing the site's cookies from
	a browser for logins a form can't do, and checking or forgetting it.
-->
<script lang="ts">
	import {
		checkAccount,
		importAccountCookies,
		loginAccount,
		logoutAccount,
		type PluginAccount,
		type PluginLogin
	} from '$lib/plugins';

	interface Props {
		plugin: string;
		login: PluginLogin;
		account: PluginAccount | null;
		onchanged: () => void;
	}

	let { plugin, login, account, onchanged }: Props = $props();

	// What each status tells the user.
	const STATUS_TEXT: Record<PluginAccount['status'], string> = {
		ok: 'Logged in',
		expired: 'The site ended the session. Log in again.',
		needsLogin: 'The stored login stopped working. Log in again.',
		failed: "Couldn't tell whether the login works."
	};

	let values = $state<Record<string, string>>({});
	let cookies = $state('');
	let busy = $state(false);
	let message = $state<string | null>(null);
	let importing = $state(false);

	/** Runs an account action, showing its failure and refreshing after. */
	async function run(action: () => Promise<unknown>, done: string) {
		busy = true;
		message = null;
		try {
			await action();
			message = done;
			onchanged();
		} catch (failure) {
			message = failure instanceof Error ? failure.message : String(failure);
		} finally {
			busy = false;
		}
	}

	/** Logs in with the form's values. */
	function submitLogin(event: SubmitEvent) {
		event.preventDefault();
		void run(() => loginAccount(plugin, values), 'Logged in.');
	}

	/** Imports the pasted cookies. */
	function submitCookies(event: SubmitEvent) {
		event.preventDefault();
		void run(async () => {
			await importAccountCookies(plugin, cookies);
			cookies = '';
			importing = false;
		}, 'Cookies imported.');
	}

	/** Reads a cookies.txt or JSON export into the field. */
	async function readCookieFile(event: Event) {
		const file = (event.currentTarget as HTMLInputElement).files?.[0];
		if (!file) return;
		cookies = await file.text();
	}
</script>

<section class="flex max-w-md flex-col gap-2">
	<h3 class="text-sm font-semibold">Account</h3>
	{#if account}
		<p class="text-sm">
			<span class:text-error={account.status !== 'ok'}>{STATUS_TEXT[account.status]}</span>
			{#if account.status === 'ok'}
				as {account.label}
			{/if}
		</p>
		<p class="text-base-content/50 text-xs">
			{#if account.storesLogin}
				Logs in again by itself when the site ends the session.
			{:else}
				Imported cookies: import them again once the site ends the session.
			{/if}
		</p>
		<div class="flex gap-2">
			<button
				type="button"
				class="bg-base-300 rounded-lg px-3 py-1.5 text-sm disabled:opacity-50"
				disabled={busy}
				onclick={() => account && run(() => checkAccount(account.id), 'Checked.')}>Check</button
			>
			<button
				type="button"
				class="text-error rounded-lg px-3 py-1.5 text-sm disabled:opacity-50"
				disabled={busy}
				onclick={() => account && run(() => logoutAccount(account.id), 'Logged out.')}
				>Log out</button
			>
		</div>
	{:else}
		<p class="text-base-content/50 text-xs">Not logged in on this device.</p>
	{/if}

	{#if login.fields.length > 0}
		<form class="flex flex-col gap-2" onsubmit={submitLogin}>
			{#each login.fields as field (field.key)}
				<label class="flex flex-col gap-1 text-sm">
					<span>{field.label}{field.required ? ' *' : ''}</span>
					<input
						type={field.type === 'password' ? 'password' : 'text'}
						autocomplete={field.type === 'password' ? 'current-password' : 'username'}
						bind:value={values[field.key]}
						required={field.required}
						class="bg-base-300 h-9 rounded-lg px-3 outline-none"
					/>
					{#if field.description}<span class="text-base-content/50 text-xs"
							>{field.description}</span
						>{/if}
				</label>
			{/each}
			<button
				type="submit"
				class="bg-primary text-primary-content h-9 self-start rounded-lg px-4 text-sm font-medium disabled:opacity-50"
				disabled={busy}>Log in</button
			>
		</form>
	{/if}

	{#if login.cookies}
		{#if importing}
			<form class="flex flex-col gap-2" onsubmit={submitCookies}>
				<label class="flex flex-col gap-1 text-sm">
					<span>Cookies</span>
					<textarea
						bind:value={cookies}
						rows="4"
						placeholder="cookies.txt, a JSON export, or a Cookie: line"
						class="bg-base-300 rounded-lg px-3 py-2 font-mono text-xs outline-none"
					></textarea>
					<span class="text-base-content/50 text-xs">
						Export them from your browser while logged into the site, e.g. with a cookies.txt or
						Cookie-Editor extension. Only this site's cookies are kept.
					</span>
				</label>
				<input
					type="file"
					accept=".txt,.json,text/plain,application/json"
					onchange={readCookieFile}
					class="text-xs"
				/>
				<div class="flex gap-2">
					<button
						type="submit"
						class="bg-primary text-primary-content h-9 rounded-lg px-4 text-sm font-medium disabled:opacity-50"
						disabled={busy || cookies.trim() === ''}>Import</button
					>
					<button type="button" class="rounded-lg px-3 text-sm" onclick={() => (importing = false)}
						>Cancel</button
					>
				</div>
			</form>
		{:else}
			<button
				type="button"
				class="bg-base-300 self-start rounded-lg px-3 py-1.5 text-sm"
				onclick={() => (importing = true)}>Import cookies from a browser</button
			>
		{/if}
	{/if}

	{#if message}<p class="text-base-content/60 text-xs">{message}</p>{/if}
</section>
