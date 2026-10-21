// Package importer imports activity files for a user. It is shared by the
// HTTP upload endpoint and the batch import command.
package importer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jlindbergh/my-tracks/backend/internal/activity"
	"github.com/jlindbergh/my-tracks/backend/internal/gpxfile"
	"github.com/jlindbergh/my-tracks/backend/internal/store"
)

// MaxFileSize is the largest file accepted for import.
const MaxFileSize = 50 << 20

// AcceptedExtensions returns the extensions of the files that can be imported.
func AcceptedExtensions() []string {
	return []string{".gpx"}
}

// Accepts reports whether filename has one of the AcceptedExtensions,
// ignoring case.
func Accepts(filename string) bool {
	ext := filepath.Ext(filename)
	return slices.ContainsFunc(AcceptedExtensions(), func(e string) bool { return strings.EqualFold(e, ext) })
}

// Status is the outcome of importing one file.
type Status string

const (
	// StatusImported means the file was stored as a new activity.
	StatusImported Status = "imported"
	// StatusAlreadyImported means the user has imported this exact file before.
	StatusAlreadyImported Status = "already_imported"
	// StatusPossibleDuplicate means the file matches an existing activity by
	// the DuplicateRule and was not stored. Import it again with
	// Options.AllowDuplicates to store it anyway.
	StatusPossibleDuplicate Status = "possible_duplicate"
	// StatusNoGPSTrack means the file was read but has no GPS track (e.g. an
	// indoor or treadmill activity), so there is nothing to show on the map.
	// Activity holds its name and sport.
	StatusNoGPSTrack Status = "no_gps_track"
	// StatusFailed means the file could not be read.
	StatusFailed Status = "failed"
)

// Result describes what happened to one file.
type Result struct {
	File       string `json:"file"`
	Status     Status `json:"status"`
	ActivityID string `json:"activityId,omitempty"`
	Error      string `json:"error,omitempty"`
	// Activity is the activity read from the file.
	Activity *activity.Facts `json:"activity,omitempty"`
	// Existing is the stored activity this file matched, for
	// already_imported and possible_duplicate.
	Existing *activity.Facts `json:"existing,omitempty"`
	// Comparison is how Activity and Existing differ, for possible_duplicate.
	Comparison *activity.Comparison `json:"comparison,omitempty"`
}

// Options control an import.
type Options struct {
	// AllowDuplicates stores possible duplicates instead of reporting them.
	// Files that were already imported are skipped regardless.
	AllowDuplicates bool
}

// Importer stores imported activities.
type Importer struct {
	store *store.Store
}

// New returns an Importer that writes to s.
func New(s *store.Store) *Importer {
	return &Importer{store: s}
}

// ImportGPX parses a GPX file and stores it as an activity owned by userID.
// Parse problems are reported in the Result; only failures to talk to the
// database are returned as errors.
func (im *Importer) ImportGPX(ctx context.Context, userID, filename string, data []byte, opts Options) (Result, error) {
	res := Result{File: filename}

	if len(data) > MaxFileSize {
		res.Status, res.Error = StatusFailed, "file too large"
		return res, nil
	}

	name := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	act, err := gpxfile.Parse(data, name)
	var noTrack *gpxfile.NoTrackError
	if errors.As(err, &noTrack) {
		res.Status = StatusNoGPSTrack
		res.Activity = &activity.Facts{Name: noTrack.Name, Sport: noTrack.Sport}
		return res, nil
	}
	if err != nil {
		res.Status, res.Error = StatusFailed, err.Error()
		return res, nil
	}

	rule := &DuplicateRule
	if opts.AllowDuplicates {
		rule = nil
	}

	sum := sha256.Sum256(data)
	ins, err := im.store.InsertActivity(ctx, userID, store.NewActivity{
		Activity:       act,
		SourceFilename: filepath.Base(filename),
		ContentSHA256:  hex.EncodeToString(sum[:]),
	}, rule)
	if err != nil {
		return Result{}, err
	}

	res.Activity, res.Existing, res.Comparison = &ins.New, ins.Existing, ins.Comparison
	switch ins.Status {
	case store.Inserted:
		res.Status, res.ActivityID = StatusImported, ins.New.ID
	case store.SameFile:
		res.Status = StatusAlreadyImported
	case store.PossibleDuplicate:
		res.Status = StatusPossibleDuplicate
	default:
		return Result{}, fmt.Errorf("unexpected insert status %v", ins.Status)
	}
	return res, nil
}
