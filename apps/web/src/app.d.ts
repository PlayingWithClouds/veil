// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces

// The dayjs plugins the root layout registers, so their methods type-check everywhere.
/// <reference types="dayjs/plugin/duration" />
/// <reference types="dayjs/plugin/relativeTime" />

declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
