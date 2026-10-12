import { writable } from 'svelte/store';

// The search being run. Home searches whenever it changes, so it is only set
// when a search is submitted (Enter, a suggestion, the search button), never
// while typing; the top bar keeps what is being typed in its own field.
export const searchQuery = writable('');
