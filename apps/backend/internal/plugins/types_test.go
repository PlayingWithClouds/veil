package plugins

import "testing"

func TestServesURL(t *testing.T) {
	plugin := &Plugin{Meta: PluginMeta{Domains: []string{"eporner.com", "*"}}}
	cases := map[string]bool{
		"https://www.eporner.com/channel/vixen/": true,
		"https://static.eporner.com/logo.png":    true,
		"https://noteporner.com/":                false,
		"https://xhamster.com/channels/vixen":    false,
		"not a url":                              false,
	}
	for rawURL, want := range cases {
		if got := plugin.ServesURL(rawURL); got != want {
			t.Errorf("ServesURL(%q) = %v, want %v", rawURL, got, want)
		}
	}
}
