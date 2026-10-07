package ingest

import "strings"

// noiseTags are listing labels sites attach to everything ("Uncategorized",
// "Sex", "HD"): they say nothing about a scene, so they never become tags.
// Keep in sync with the list in migration 006_merge_tag_case.sql.
var noiseTags = map[string]bool{
	"uncategorized": true, "other": true, "others": true,
	"sex": true, "porn": true, "porno": true, "xxx": true,
	"video": true, "videos": true, "hd": true, "hd porn": true, "free porn": true,
	"porn videos": true, "sex videos": true, "full video": true, "full movie": true,
}

// withoutNoiseTags drops the tag names that carry no information.
func withoutNoiseTags(names []string) []string {
	kept := make([]string, 0, len(names))
	for _, name := range names {
		if noiseTags[strings.ToLower(strings.TrimSpace(name))] {
			continue
		}
		kept = append(kept, name)
	}
	return kept
}
