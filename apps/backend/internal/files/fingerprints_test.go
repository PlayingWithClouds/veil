package files

import (
	"bytes"
	"context"
	"slices"
	"testing"

	"github.com/playingwithclouds/veil/internal/schema"
)

// TestOshash checks the hash against a hand-computed value: an 8-byte file
// holding 1 is read as head and tail (1 + 1) plus its size (8).
func TestOshash(t *testing.T) {
	content := []byte{1, 0, 0, 0, 0, 0, 0, 0}
	hash, err := computeOshash(bytes.NewReader(content), int64(len(content)))
	if err != nil || hash != "000000000000000a" {
		t.Errorf("oshash = %s, err %v", hash, err)
	}
}

// TestSameMedia finds records sharing an exact fingerprint or a phash a few
// bits away, and leaves distant phashes out.
func TestSameMedia(t *testing.T) {
	ctx := context.Background()
	env := newTestEnv(t)
	first := env.newVideo(t, "First")
	near := env.newVideo(t, "Near")
	far := env.newVideo(t, "Far")
	exact := env.newVideo(t, "Exact")
	add := func(subject schema.RecordRef, algorithm schema.FingerprintAlgorithm, value string) *schema.Fingerprint {
		fingerprint, err := env.service.AddFingerprint(ctx, subject, nil, algorithm, value, nil)
		if err != nil {
			t.Fatal(err)
		}
		return fingerprint
	}
	stored := add(first, schema.FingerprintAlgorithmPhash, "FFFFFFFFFFFFFFFF")
	add(near, schema.FingerprintAlgorithmPhash, "fffffffffffffff0")
	add(far, schema.FingerprintAlgorithmPhash, "0000000000000000")
	add(first, schema.FingerprintAlgorithmMd5, "0123456789abcdef0123456789abcdef")
	add(exact, schema.FingerprintAlgorithmMd5, "0123456789abcdef0123456789abcdef")

	if again := add(first, schema.FingerprintAlgorithmPhash, "ffffffffffffffff"); again.Id != stored.Id {
		t.Error("the same fingerprint was stored twice")
	}
	same, err := env.service.SameMedia(ctx, first)
	if err != nil {
		t.Fatal(err)
	}
	if len(same) != 2 || !slices.Contains(same, near) || !slices.Contains(same, exact) {
		t.Errorf("same = %+v, want near and exact", same)
	}
	if _, err := env.service.AddFingerprint(ctx, first, nil, schema.FingerprintAlgorithmPhash, "xyz", nil); err == nil {
		t.Error("invalid phash accepted")
	}
	if _, err := env.service.AddFingerprint(ctx, first, nil, "sha1", "abc", nil); err == nil {
		t.Error("unknown algorithm accepted")
	}
}

// TestFingerprintTask recomputes a file's fingerprints through a job.
func TestFingerprintTask(t *testing.T) {
	ctx := context.Background()
	env := newTestEnv(t)
	root, folder := env.libraryRoot(t, nil)
	writeFile(t, folder+"/clip.mp4", patterned(1024, 1))
	env.scan(t, root.Id)
	clip := addedAt(t, env.added, "media/clip.mp4")
	if _, err := env.service.FingerprintFile(ctx, clip.Id); err != nil {
		t.Fatal(err)
	}
	env.runJob(t, schema.JobKindFingerprint, env.service.FingerprintTask())
	fingerprints, _ := env.service.Fingerprints(ctx, schema.RecordRef{Kind: clip.SubjectKind, Id: clip.SubjectId})
	if len(fingerprints) != 2 {
		t.Errorf("fingerprints = %+v, want oshash and md5 once each", fingerprints)
	}
}
