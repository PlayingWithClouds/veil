package accounts

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// netscapeFields is the column count of a cookies.txt line: domain, include
// subdomains, path, secure, expires, name, value.
const netscapeFields = 7

// browserCookie is one cookie as browser extensions export them
// (Cookie-Editor, EditThisCookie): expirationDate in seconds, maybe
// fractional; some exports say expires or expiry instead.
type browserCookie struct {
	Name           string   `json:"name"`
	Value          string   `json:"value"`
	Domain         string   `json:"domain"`
	Path           string   `json:"path"`
	Secure         bool     `json:"secure"`
	ExpirationDate *float64 `json:"expirationDate"`
	Expires        *float64 `json:"expires"`
	Expiry         *float64 `json:"expiry"`
}

// ParseCookies reads cookies exported from a browser: a JSON export (an
// array, or an object with a cookies array), a cookies.txt file, or a Cookie
// header ("a=1; b=2", which says no domain).
func ParseCookies(text string) ([]plugins.Cookie, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("no cookies given")
	}
	if strings.HasPrefix(text, "[") || strings.HasPrefix(text, "{") {
		return parseJSONCookies(text)
	}
	if strings.Contains(text, "\t") {
		return parseNetscapeCookies(text), nil
	}
	return parseCookieHeader(text), nil
}

// parseJSONCookies reads a browser extension's JSON export.
func parseJSONCookies(text string) ([]plugins.Cookie, error) {
	var exported []browserCookie
	if strings.HasPrefix(text, "{") {
		var wrapped struct {
			Cookies []browserCookie `json:"cookies"`
		}
		if err := json.Unmarshal([]byte(text), &wrapped); err != nil {
			return nil, fmt.Errorf("cookies aren't valid JSON: %w", err)
		}
		exported = wrapped.Cookies
	} else if err := json.Unmarshal([]byte(text), &exported); err != nil {
		return nil, fmt.Errorf("cookies aren't valid JSON: %w", err)
	}
	cookies := make([]plugins.Cookie, 0, len(exported))
	for _, cookie := range exported {
		if cookie.Name == "" {
			continue
		}
		cookies = append(cookies, plugins.Cookie{
			Name: cookie.Name, Value: cookie.Value, Domain: cookie.Domain, Path: cookie.Path,
			Secure: cookie.Secure, Expires: cookie.expiry(),
		})
	}
	return cookies, nil
}

// expiry is the cookie's expiry in whole seconds, 0 for a session cookie.
func (cookie browserCookie) expiry() int64 {
	for _, seconds := range []*float64{cookie.ExpirationDate, cookie.Expires, cookie.Expiry} {
		if seconds != nil {
			return int64(*seconds)
		}
	}
	return 0
}

// parseNetscapeCookies reads a cookies.txt file; comments, blank and
// malformed lines are skipped.
func parseNetscapeCookies(text string) []plugins.Cookie {
	var cookies []plugins.Cookie
	for _, line := range strings.Split(text, "\n") {
		cookie, valid := netscapeCookie(line)
		if valid {
			cookies = append(cookies, cookie)
		}
	}
	return cookies
}

// netscapeCookie reads one cookies.txt line. HttpOnly cookies are written
// with a "#HttpOnly_" prefix, which isn't a comment.
func netscapeCookie(line string) (plugins.Cookie, bool) {
	line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "#HttpOnly_"))
	if line == "" || strings.HasPrefix(line, "#") {
		return plugins.Cookie{}, false
	}
	fields := strings.Split(line, "\t")
	if len(fields) < netscapeFields {
		return plugins.Cookie{}, false
	}
	expires, _ := strconv.ParseInt(fields[4], 10, 64)
	return plugins.Cookie{
		Domain: fields[0], Path: fields[2], Secure: strings.EqualFold(fields[3], "TRUE"),
		Expires: expires, Name: fields[5], Value: fields[6],
	}, fields[5] != ""
}

// parseCookieHeader reads "a=1; b=2", with or without "Cookie:" before it.
func parseCookieHeader(text string) []plugins.Cookie {
	text = strings.TrimSpace(strings.TrimPrefix(text, "Cookie:"))
	var cookies []plugins.Cookie
	for _, pair := range strings.Split(text, ";") {
		name, value, found := strings.Cut(strings.TrimSpace(pair), "=")
		if found && name != "" {
			cookies = append(cookies, plugins.Cookie{Name: name, Value: value})
		}
	}
	return cookies
}

// cookiesFor keeps the cookies of a plugin's domains, so a whole browser
// export doesn't hand the plugin every site's login. A cookie without a
// domain (from a Cookie header) is taken for the plugin's first domain.
func cookiesFor(cookies []plugins.Cookie, domains []string) []plugins.Cookie {
	siteDomains := ownDomains(domains)
	if len(siteDomains) == 0 {
		return nil
	}
	var kept []plugins.Cookie
	for _, cookie := range cookies {
		if cookie.Domain == "" {
			cookie.Domain = siteDomains[0]
		}
		if belongsTo(cookie.Domain, siteDomains) {
			kept = append(kept, cookie)
		}
	}
	return kept
}

// ownDomains are a plugin's domains without the wildcard.
func ownDomains(domains []string) []string {
	var own []string
	for _, domain := range domains {
		if domain != "*" && domain != "" {
			own = append(own, strings.ToLower(domain))
		}
	}
	return own
}

// belongsTo reports whether a cookie's domain is one of the site's, a
// subdomain of one, or a parent domain one of them sends cookies to. A bare
// top-level domain ("com") is nobody's.
func belongsTo(cookieDomain string, domains []string) bool {
	cookieDomain = strings.ToLower(strings.TrimPrefix(cookieDomain, "."))
	if !strings.Contains(cookieDomain, ".") {
		return false
	}
	for _, domain := range domains {
		if cookieDomain == domain || strings.HasSuffix(cookieDomain, "."+domain) || strings.HasSuffix(domain, "."+cookieDomain) {
			return true
		}
	}
	return false
}
