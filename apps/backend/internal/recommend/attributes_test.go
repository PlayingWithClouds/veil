package recommend

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
)

func TestAttributeKeysReadTheProfileColumns(t *testing.T) {
	row := db.Row{
		"hair_color": "Brown", "ethnicity": "Caucasian", "eye_color": "Unknown", "country": "United States",
		"fake_tits": "Fake", "tattoos": "Yes: Under left bicep", "piercings": "", "measurements": "86DD-63-86 cm",
		"height_cm": int64(157),
	}
	got := attributeKeys(row)
	want := []string{
		"hair:brown", "ethnicity:caucasian", "country:united states", "boobs:fake",
		"tattoos:yes", "cup:large", "height:short",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("keys = %v, want %v", got, want)
	}
}

func TestCupAndHeatBandsAreCoarse(t *testing.T) {
	cups := map[string]string{
		"81B cm": "small", "32A-24-35": "small", "93H-58-83 cm": "large", "81C-60-83 cm": "medium",
		"86DD-63-86 cm": "large", "86 cm": "", "": "", "Natural": "",
	}
	for measurements, want := range cups {
		if got := cupBand(measurements); got != want {
			t.Errorf("cupBand(%q) = %q, want %q", measurements, got, want)
		}
	}
	heights := map[int]string{0: "", 150: "short", 165: "average", 175: "tall"}
	for height, want := range heights {
		if got := heightBand(height); got != want {
			t.Errorf("heightBand(%d) = %q, want %q", height, got, want)
		}
	}
}

// testAttributes indexes three performers: two redheads and one brunette.
func testAttributes() *performerAttributes {
	attributes := &performerAttributes{byPerformer: map[string][]string{
		"performer:liked":    {"hair:red", "eyes:green"},
		"performer:twin":     {"hair:red", "eyes:green"},
		"performer:stranger": {"hair:brown", "eyes:brown"},
		"performer:disliked": {"hair:brown"},
	}, frequency: map[string]int{}}
	for _, keys := range attributes.byPerformer {
		for _, key := range keys {
			attributes.frequency[key]++
		}
	}
	return attributes
}

func TestAttributeAffinitiesSpreadPerformerAffinityOntoLooks(t *testing.T) {
	performers := map[string]float64{"performer:liked": 1, "performer:disliked": -1, "performer:blocked": -1}
	blocked := map[string]bool{"performer:blocked": true}
	attributes := testAttributes()
	attributes.byPerformer["performer:blocked"] = []string{"hair:blonde"}

	got := attributeAffinities(performers, blocked, attributes)
	if got["hair:red"] <= 0 || got["hair:brown"] >= 0 {
		t.Errorf("liked looks must be positive, disliked negative: %v", got)
	}
	if _, present := got["hair:blonde"]; present {
		t.Error("a blocked performer's looks must not count")
	}
	for key, value := range got {
		if value < -1 || value > 1 {
			t.Errorf("%s = %v out of range", key, value)
		}
	}
	// One liked performer is thin evidence: far from the full ±1.
	if got["eyes:green"] >= 0.5 {
		t.Errorf("one performer should leave a modest affinity, got %v", got["eyes:green"])
	}
}

func TestUnknownPerformersInheritDampedAffinityFromLooks(t *testing.T) {
	attributes := testAttributes()
	profile := newProfile(1)
	profile.attributes = attributes
	profile.Performers = map[string]float64{"performer:liked": 1, "performer:disliked": -1}
	profile.Attributes = attributeAffinities(profile.Performers, nil, attributes)

	twin := castAffinity([]string{"performer:twin"}, profile)
	stranger := castAffinity([]string{"performer:stranger"}, profile)
	if twin <= 0 || twin >= profile.Performers["performer:liked"] {
		t.Errorf("the twin inherits less than the original: %v", twin)
	}
	if stranger >= 0 {
		t.Errorf("a performer who looks like a disliked one inherits the dislike, got %v", stranger)
	}
	assertClose(t, "known performer keeps her own affinity", castAffinity([]string{"performer:liked"}, profile), 1)
	assertClose(t, "no attributes, no opinion", castAffinity([]string{"performer:nobody"}, profile), 0)
	if best := castAffinity([]string{"performer:stranger", "performer:twin"}, profile); best != twin {
		t.Errorf("the strongest of the cast carries the scene: got %v, want %v", best, twin)
	}
}

func TestAttributeCandidatesProposeScenesOfLookAlikes(t *testing.T) {
	ctx := context.Background()
	engine, database, _ := newTestEngine(t)
	for name, fields := range map[string]map[string]any{
		"performer:liked":    {"name": "Liked", "hair_color": "Red", "eye_color": "Green"},
		"performer:twin":     {"name": "Twin", "hair_color": "Red", "eye_color": "Green"},
		"performer:stranger": {"name": "Stranger", "hair_color": "Black", "eye_color": "Brown"},
	} {
		insertRow(t, database, "performer", name, fields)
	}
	// Enough other performers for red hair and green eyes to be rare looks.
	for index := 0; index < 12; index++ {
		insertRow(t, database, "performer", fmt.Sprintf("performer:filler%d", index), map[string]any{"name": "Filler", "hair_color": "Blonde", "eye_color": "Blue"})
	}
	insertScene(t, database, "scene:liked", map[string]any{"performers": []string{"performer:liked"}})
	insertScene(t, database, "scene:twin", map[string]any{"performers": []string{"performer:twin"}})
	insertScene(t, database, "scene:filler", map[string]any{"performers": []string{"performer:filler0"}})
	insertScene(t, database, "scene:stranger", map[string]any{"performers": []string{"performer:stranger"}})
	insertRow(t, database, "user_rating", "user_rating:1", map[string]any{"media": "scene:liked", "rating": 10, "updated_at": ago(0)})
	insertRow(t, database, "user_rating", "user_rating:2", map[string]any{"media": "scene:stranger", "rating": 1, "updated_at": ago(0)})

	state, err := engine.takeSnapshot(ctx, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	found, err := engine.attributeCandidates(ctx, state)
	if err != nil {
		t.Fatalf("attribute candidates: %v", err)
	}
	if len(found) != 1 || found[0].sceneID != "scene:twin" || found[0].source != SourceExploreAttribute {
		t.Fatalf("candidates = %+v, want only the twin's scene", found)
	}
	if found[0].reason.Kind != ReasonPerformer || found[0].reason.EntityID != "performer:twin" {
		t.Errorf("reason = %+v", found[0].reason)
	}
	if sourceKind(found[0].source) != kindExplore {
		t.Error("attribute finds are exploration")
	}

	items := feed(t, engine, nil)
	twin := itemBySceneID(items, "scene:twin")
	filler := itemBySceneID(items, "scene:filler")
	if twin == nil || filler == nil || twin.Score <= filler.Score {
		t.Errorf("the look-alike must outscore an unrelated unknown performer's scene: %+v / %+v", twin, filler)
	}
}
