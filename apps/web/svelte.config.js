import adapterNode from '@sveltejs/adapter-node';
import adapterStatic from '@sveltejs/adapter-static';

/**
 * Node server by default; `VEIL_TARGET=mobile` builds a static SPA into the
 * Capacitor app's web dir instead (apps/mobile/www).
 */
function pickAdapter() {
	if (process.env.VEIL_TARGET === 'mobile') {
		return adapterStatic({
			pages: '../mobile/www',
			assets: '../mobile/www',
			fallback: 'index.html'
		});
	}
	return adapterNode();
}

/** @type {import('@sveltejs/kit').Config} */
const config = {
	compilerOptions: {
		runes: ({ filename }) => (filename.split(/[/\\]/).includes('node_modules') ? undefined : true)
	},
	kit: {
		adapter: pickAdapter()
	}
};

export default config;
