import type { CapacitorConfig } from '@capacitor/cli';

// The web client (apps/web) built as a static SPA into www/. It talks to a
// backend the user enters on first launch, usually over plain http on the LAN,
// so the app is served from http://localhost (not https) and cleartext is on:
// otherwise every request to an http backend is blocked as mixed content.
const config: CapacitorConfig = {
	appId: 'com.playingwithclouds.veil',
	appName: 'Veil',
	webDir: 'www',
	plugins: {
		SplashScreen: {
			launchShowDuration: 2000,
			launchAutoHide: true,
			backgroundColor: '#0d0b1aff',
			androidSplashResourceName: 'splash',
			androidScaleType: 'CENTER_CROP',
			showSpinner: false
		}
	},
	server: {
		androidScheme: 'http',
		cleartext: true
	}
};

// Live reload: point the app at a running web dev server instead of www/.
if (process.env.CAP_SERVER_URL) {
	config.server!.url = process.env.CAP_SERVER_URL;
}

export default config;
