package alike

const (
	// sameImageSimilarity is how alike two difference hashes must be (at
	// most 6 of 64 bits apart) to be the same picture re-encoded, rescaled or
	// watermarked by another site.
	sameImageSimilarity = 0.9

	// sameFramesMinimum is how many of the sites' stills across the video
	// must match for the frames alone to show the same video.
	sameFramesMinimum = 2

	// sameByImagesDurationRatio is how close the runtimes must be when the
	// pictures are the evidence: stricter than the title rule, since one
	// studio's cover can front both a scene and its longer cut.
	sameByImagesDurationRatio = 0.98
)

// Pictures are a video's image hashes: its cover's and those of the stills
// the site shows across it (preview frames). Zero hashes are unknown.
type Pictures struct {
	Cover  uint64
	Frames []uint64
}

// SameImage reports whether two difference hashes are the same picture.
func SameImage(first, second uint64) bool {
	return HammingSimilarity(first, second) >= sameImageSimilarity
}

// matchingFrames counts the frames of first that have a match among second's.
func matchingFrames(first, second []uint64) int {
	count := 0
	for _, frame := range first {
		for _, other := range second {
			if SameImage(frame, other) {
				count++
				break
			}
		}
	}
	return count
}

// PicturesAgree reports whether two videos show the same pictures: the same
// cover, or at least sameFramesMinimum of the same stills. A cover the
// caller knows to be a placeholder (shared by unrelated videos) doesn't
// count, so coverUsable says whether the covers may be compared.
func PicturesAgree(first, second Pictures, coverUsable bool) bool {
	if coverUsable && SameImage(first.Cover, second.Cover) {
		return true
	}
	return matchingFrames(first.Frames, second.Frames) >= sameFramesMinimum
}

// SameSceneByImages reports whether two videos are provably one scene from
// their pictures: they agree (PicturesAgree), both runtimes are known and
// within sameByImagesDurationRatio, and neither the titles' numbers nor the
// casts tell them apart. Strict enough to merge on, since a wrong merge loses
// a scene; titles may differ entirely, as sites reword and translate them.
func SameSceneByImages(first, second Candidate, firstPictures, secondPictures Pictures, coverUsable bool) bool {
	if !runtimesWithin(first.DurationSeconds, second.DurationSeconds, sameByImagesDurationRatio) {
		return false
	}
	if conflictingNumbers(titleTokens(first.Title), titleTokens(second.Title)) || castsConflict(first, second) {
		return false
	}
	return PicturesAgree(firstPictures, secondPictures, coverUsable)
}

// runtimesWithin reports whether both runtimes are known and the shorter is
// at least ratio of the longer.
func runtimesWithin(firstSeconds, secondSeconds int, ratio float64) bool {
	if firstSeconds <= 0 || secondSeconds <= 0 {
		return false
	}
	shorter, longer := firstSeconds, secondSeconds
	if shorter > longer {
		shorter, longer = longer, shorter
	}
	return float64(shorter)/float64(longer) >= ratio
}
