import { writable } from 'svelte/store';

// The query in the global top bar. Explore searches as it changes; other pages
// hand it to Explore on Enter.
export const searchQuery = writable('');
