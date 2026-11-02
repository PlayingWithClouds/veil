package provenance

import "github.com/playingwithclouds/veil/internal/schema"

// Automation is the author of the changes one automation run writes. Its
// revisions carry the run id, so undoing the run reverts exactly them.
func Automation(name, runID string) Author {
	return Author{Actor: schema.ActorAutomation, Name: name, AutomationRunID: runID}
}
