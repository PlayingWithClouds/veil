package files

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"math/bits"
	"slices"
	"strconv"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/schema"
)

// maxPhashDistance is how many of a perceptual hash's 64 bits may differ for
// two videos to count as the same media.
const maxPhashDistance = 8

// algorithms are the valid fingerprint algorithms.
var algorithms = []schema.FingerprintAlgorithm{
	schema.FingerprintAlgorithmPhash,
	schema.FingerprintAlgorithmOshash,
	schema.FingerprintAlgorithmMd5,
	schema.FingerprintAlgorithmAudio,
}

// computedAlgorithms are the fingerprints the backend computes from a file;
// the others come from clients.
var computedAlgorithms = []schema.FingerprintAlgorithm{schema.FingerprintAlgorithmOshash, schema.FingerprintAlgorithmMd5}

// AddFingerprint stores a fingerprint of a record, typically one a client
// computed (phash and audio need ffmpeg, which the backend lacks). fileID
// names the file it was computed from, nil for a stream. Submitting the same
// fingerprint again returns the stored one.
func (s *Service) AddFingerprint(ctx context.Context, subject schema.RecordRef, fileID *string, algorithm schema.FingerprintAlgorithm, value string, durationSeconds *float64) (*schema.Fingerprint, error) {
	value, err := normalizeFingerprint(algorithm, value)
	if err != nil {
		return nil, err
	}
	if err := s.requireRecord(ctx, subject); err != nil {
		return nil, err
	}
	vars := db.Vars{"kind": subject.Kind, "id": subject.Id, "algorithm": algorithm, "value": value, "file": fileID}
	existing, err := db.QueryOneAs[schema.Fingerprint](ctx, s.database,
		`SELECT * FROM "Fingerprint" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "algorithm" = $algorithm
		 AND "value" = $value AND "fileId" IS $file AND "deletedAt" IS NULL`, vars)
	if err != nil || existing != nil {
		return existing, err
	}
	id, err := s.database.Insert(ctx, tableFingerprint, schema.Fingerprint{
		SubjectKind: subject.Kind, SubjectId: subject.Id, FileId: fileID,
		Algorithm: algorithm, Value: value, DurationSeconds: durationSeconds,
	})
	if err != nil {
		return nil, err
	}
	return db.Get[schema.Fingerprint](ctx, s.database, tableFingerprint, id)
}

// normalizeFingerprint checks a submitted value and returns it in stored
// form: hex hashes lowercase, a phash exactly 64 bits.
func normalizeFingerprint(algorithm schema.FingerprintAlgorithm, value string) (string, error) {
	if !slices.Contains(algorithms, algorithm) {
		return "", fmt.Errorf("unknown fingerprint algorithm %q", algorithm)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("a fingerprint needs a value")
	}
	if algorithm == schema.FingerprintAlgorithmAudio {
		return value, nil
	}
	value = strings.ToLower(value)
	digits := 16
	if algorithm == schema.FingerprintAlgorithmMd5 {
		digits = 32
	}
	if _, err := hex.DecodeString(value); err != nil || len(value) != digits {
		return "", fmt.Errorf("a %s is %d hex digits, not %q", algorithm, digits, value)
	}
	return value, nil
}

// Fingerprints lists a record's fingerprints.
func (s *Service) Fingerprints(ctx context.Context, subject schema.RecordRef) ([]schema.Fingerprint, error) {
	return db.QueryAs[schema.Fingerprint](ctx, s.database,
		`SELECT * FROM "Fingerprint" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "deletedAt" IS NULL
		 ORDER BY "algorithm", "createdAt"`, db.Vars{"kind": subject.Kind, "id": subject.Id})
}

// SameMedia lists the other records that share a fingerprint with subject:
// an equal value of the same algorithm, or for phash one at most
// maxPhashDistance bits away.
func (s *Service) SameMedia(ctx context.Context, subject schema.RecordRef) ([]schema.RecordRef, error) {
	vars := db.Vars{"kind": subject.Kind, "id": subject.Id, "phash": schema.FingerprintAlgorithmPhash}
	exact, err := db.QueryAs[schema.RecordRef](ctx, s.database,
		`SELECT DISTINCT other."subjectKind" AS "kind", other."subjectId" AS "id"
		 FROM "Fingerprint" own JOIN "Fingerprint" other ON other."algorithm" = own."algorithm" AND other."value" = own."value"
		 WHERE own."subjectKind" = $kind AND own."subjectId" = $id AND own."algorithm" != $phash
		   AND own."deletedAt" IS NULL AND other."deletedAt" IS NULL
		   AND NOT (other."subjectKind" = $kind AND other."subjectId" = $id)`, vars)
	if err != nil {
		return nil, err
	}
	near, err := s.nearPhashes(ctx, subject)
	if err != nil {
		return nil, err
	}
	return uniqueRefs(append(exact, near...)), nil
}

// nearPhashes lists the other records with a phash close to one of subject's.
func (s *Service) nearPhashes(ctx context.Context, subject schema.RecordRef) ([]schema.RecordRef, error) {
	phashes, err := db.QueryAs[schema.Fingerprint](ctx, s.database,
		`SELECT * FROM "Fingerprint" WHERE "algorithm" = $phash AND "deletedAt" IS NULL`,
		db.Vars{"phash": schema.FingerprintAlgorithmPhash})
	if err != nil {
		return nil, err
	}
	own, others := splitBySubject(phashes, subject)
	var near []schema.RecordRef
	for _, other := range others {
		if isNearAny(other.Value, own) {
			near = append(near, schema.RecordRef{Kind: other.SubjectKind, Id: other.SubjectId})
		}
	}
	return near, nil
}

// splitBySubject separates subject's fingerprint values from the others'.
func splitBySubject(fingerprints []schema.Fingerprint, subject schema.RecordRef) ([]string, []schema.Fingerprint) {
	var own []string
	var others []schema.Fingerprint
	for _, fingerprint := range fingerprints {
		if fingerprint.SubjectKind == subject.Kind && fingerprint.SubjectId == subject.Id {
			own = append(own, fingerprint.Value)
			continue
		}
		others = append(others, fingerprint)
	}
	return own, others
}

// isNearAny reports whether a phash is within maxPhashDistance of any of
// candidates.
func isNearAny(value string, candidates []string) bool {
	for _, candidate := range candidates {
		distance, ok := phashDistance(value, candidate)
		if ok && distance <= maxPhashDistance {
			return true
		}
	}
	return false
}

// phashDistance is the number of differing bits of two 64-bit hex hashes;
// ok is false when one isn't a hash.
func phashDistance(first, second string) (int, bool) {
	firstBits, err := strconv.ParseUint(first, 16, 64)
	if err != nil {
		return 0, false
	}
	secondBits, err := strconv.ParseUint(second, 16, 64)
	if err != nil {
		return 0, false
	}
	return bits.OnesCount64(firstBits ^ secondBits), true
}

// uniqueRefs drops repeated refs, keeping the first.
func uniqueRefs(refs []schema.RecordRef) []schema.RecordRef {
	seen := map[schema.RecordRef]bool{}
	unique := []schema.RecordRef{}
	for _, ref := range refs {
		if seen[ref] {
			continue
		}
		seen[ref] = true
		unique = append(unique, ref)
	}
	return unique
}

// FingerprintFile queues computing a file's oshash (and md5) and returns
// the job's id.
func (s *Service) FingerprintFile(ctx context.Context, fileID string) (string, error) {
	file, err := s.File(ctx, fileID)
	if err != nil {
		return "", err
	}
	return s.config.Queue.Enqueue(ctx, jobs.Job{
		Kind:      schema.JobKindFingerprint,
		Subject:   &schema.RecordRef{Kind: file.SubjectKind, Id: file.SubjectId},
		Payload:   map[string]any{"file": fileID},
		DedupeKey: "fingerprint:" + fileID,
	})
}

// FingerprintTask runs fingerprint jobs: the payload's file is hashed and
// its computed fingerprints replaced.
func (s *Service) FingerprintTask() jobs.Task {
	return func(ctx context.Context, job *jobs.Job, progress func(float64)) error {
		fileID := payloadString(job, "file")
		if fileID == "" {
			return fmt.Errorf("a fingerprint job needs a file")
		}
		file, err := s.File(ctx, fileID)
		if err != nil {
			return err
		}
		hashes, err := s.hashFile(ctx, *file)
		if err != nil {
			return err
		}
		return s.replaceFingerprints(ctx, *file, hashes)
	}
}

// hashFile fingerprints a stored file, reading blob-store files decrypted.
func (s *Service) hashFile(ctx context.Context, file schema.MediaFile) ([]hashValue, error) {
	opened, err := s.open(ctx, file)
	if err != nil {
		return nil, err
	}
	defer opened.content.Close()
	size, err := opened.content.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	return computeHashes(opened.content, size)
}

// replaceFingerprints swaps the computed fingerprints of a file for hashes.
func (s *Service) replaceFingerprints(ctx context.Context, file schema.MediaFile, hashes []hashValue) error {
	return s.database.Tx(ctx, func(tx *db.DB) error {
		if _, err := tx.Exec(ctx,
			`DELETE FROM "Fingerprint" WHERE "fileId" = $file AND "algorithm" IN (SELECT value FROM json_each($algorithms))`,
			db.Vars{"file": file.Id, "algorithms": computedAlgorithms}); err != nil {
			return err
		}
		for _, hash := range hashes {
			row := schema.Fingerprint{
				SubjectKind: file.SubjectKind, SubjectId: file.SubjectId, FileId: &file.Id,
				Algorithm: hash.algorithm, Value: hash.value, DurationSeconds: file.DurationSeconds,
			}
			if _, err := tx.Insert(ctx, tableFingerprint, row); err != nil {
				return err
			}
		}
		return nil
	})
}
