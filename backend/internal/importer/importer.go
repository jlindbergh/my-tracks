// Package importer imports activity files for a user. It is shared by the
// HTTP upload endpoint and the batch import command.
package importer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"strings"

	"github.com/jlindbergh/my-tracks/backend/internal/gpxfile"
	"github.com/jlindbergh/my-tracks/backend/internal/store"
)

// MaxFileSize is the largest file accepted for import.
const MaxFileSize = 50 << 20

// Status is the outcome of importing one file.
type Status string

const (
	StatusImported  Status = "imported"
	StatusDuplicate Status = "duplicate"
	StatusFailed    Status = "failed"
)

// Result describes what happened to one file.
type Result struct {
	File       string `json:"file"`
	Status     Status `json:"status"`
	ActivityID string `json:"activityId,omitempty"`
	Error      string `json:"error,omitempty"`
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
func (im *Importer) ImportGPX(ctx context.Context, userID, filename string, data []byte) (Result, error) {
	res := Result{File: filename}

	if len(data) > MaxFileSize {
		res.Status, res.Error = StatusFailed, "file too large"
		return res, nil
	}

	name := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	act, err := gpxfile.Parse(data, name)
	if err != nil {
		res.Status, res.Error = StatusFailed, err.Error()
		return res, nil
	}

	sum := sha256.Sum256(data)
	id, created, err := im.store.InsertActivity(ctx, userID, store.NewActivity{
		Activity:       act,
		SourceFilename: filepath.Base(filename),
		ContentSHA256:  hex.EncodeToString(sum[:]),
	})
	if err != nil {
		return Result{}, err
	}
	if !created {
		res.Status = StatusDuplicate
		return res, nil
	}
	res.Status, res.ActivityID = StatusImported, id
	return res, nil
}
