package alike

import (
	"net/url"
	"strings"
)

// PersonFacts is what tells two performer records with a shared name or
// alias apart, or proves them one person. Empty strings and zero hashes are
// unknown.
type PersonFacts struct {
	Birthdate string
	Gender    string
	// PictureHash is the difference hash of their profile picture.
	PictureHash uint64
}

// SamePerson reports whether two performers who already share a name or an
// alias (the caller checks that) are provably one person, strict enough to
// merge on: nothing they both state may differ (birthdate, gender), and one
// thing must tie them: the same birthdate, a video crediting both (a copy of
// one scene credits each spelling), or the same profile picture. A shared
// name alone is not enough: two people can use one stage name.
func SamePerson(first, second PersonFacts, sharedVideos int) bool {
	firstBirth, secondBirth := day(first.Birthdate), day(second.Birthdate)
	if firstBirth != "" && secondBirth != "" && firstBirth != secondBirth {
		return false
	}
	if differs(first.Gender, second.Gender) {
		return false
	}
	if firstBirth != "" && firstBirth == secondBirth {
		return true
	}
	if sharedVideos > 0 {
		return true
	}
	return SameImage(first.PictureHash, second.PictureHash)
}

// StudioFacts is what proves two studio records with a shared name or alias
// one company. Empty strings and zero hashes are unknown.
type StudioFacts struct {
	URL         string
	Parent      string
	PictureHash uint64
}

// SameStudio reports whether two studios that share a name or alias are
// provably one: their sites are on the same host, they belong to the same
// network, or they show the same logo. Different hosts tell them apart.
func SameStudio(first, second StudioFacts) bool {
	firstHost, secondHost := host(first.URL), host(second.URL)
	if firstHost != "" && secondHost != "" {
		return firstHost == secondHost
	}
	if first.Parent != "" && first.Parent == second.Parent {
		return true
	}
	return SameImage(first.PictureHash, second.PictureHash)
}

// day is the YYYY-MM-DD part of a date, "" when there's none.
func day(date string) string {
	date = strings.TrimSpace(date)
	if len(date) < len("2006-01-02") {
		return ""
	}
	return date[:len("2006-01-02")]
}

// differs reports whether two stated values are both known and different,
// ignoring case.
func differs(first, second string) bool {
	first, second = strings.TrimSpace(first), strings.TrimSpace(second)
	return first != "" && second != "" && !strings.EqualFold(first, second)
}

// host is a URL's host without "www.", "" when it has none.
func host(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www.")
}
