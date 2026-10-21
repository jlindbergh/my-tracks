package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jlindbergh/my-tracks/backend/internal/activity"
	"github.com/jlindbergh/my-tracks/backend/internal/auth"
	"github.com/jlindbergh/my-tracks/backend/internal/importer"
	"github.com/jlindbergh/my-tracks/backend/internal/store"
)

const validGPX = `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg>
	<trkpt lat="59.3" lon="18.0"/><trkpt lat="59.4" lon="18.1"/>
</trkseg></trk></gpx>`

// flakyStore fails to store the files named in fail and stores the others.
type flakyStore struct {
	fail map[string]bool
}

func (s flakyStore) InsertActivity(_ context.Context, _ string, na store.NewActivity, _ *activity.DuplicateRule) (store.InsertResult, error) {
	if s.fail[na.SourceFilename] {
		return store.InsertResult{}, errors.New("connection refused")
	}
	return store.InsertResult{Status: store.Inserted, New: activity.Facts{ID: "id-" + na.SourceFilename}}, nil
}

func TestImportReportsStoreErrorsPerFile(t *testing.T) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for _, name := range []string{"a.gpx", "b.gpx", "c.gpx"} {
		fw, err := mw.CreateFormFile("files", name)
		if err != nil {
			t.Fatal(err)
		}
		io.WriteString(fw, validGPX)
	}
	mw.Close()

	im := importer.New(flakyStore{fail: map[string]bool{"b.gpx": true}})
	srv := New(nil, im, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	h := auth.DevMiddleware(store.User{ID: "u"}, srv.Routes())
	req := httptest.NewRequest(http.MethodPost, "/api/activities/import", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 with a result per file", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "connection refused") {
		t.Error("the response reveals the database error")
	}
	var results []importer.Result
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatal(err)
	}
	want := []importer.Status{importer.StatusImported, importer.StatusFailed, importer.StatusImported}
	if len(results) != len(want) {
		t.Fatalf("got %d results, want %d: %+v", len(results), len(want), results)
	}
	for i, r := range results {
		if r.Status != want[i] {
			t.Errorf("%s: status = %q, want %q", r.File, r.Status, want[i])
		}
	}
	if results[1].Error == "" {
		t.Error("the failed file has no message for the user")
	}
}
