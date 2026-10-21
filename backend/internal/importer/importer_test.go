package importer

import (
	"context"
	"strings"
	"testing"
)

func TestImportGPXReportsInvalidFiles(t *testing.T) {
	// Invalid files are rejected before anything is stored, so no store is needed.
	im := New(nil)
	for name, doc := range map[string]string{
		"not xml": "not xml",
		"invalid position": `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg>
			<trkpt lat="59.3" lon="18.0"/><trkpt lat="NaN" lon="18.1"/>
		</trkseg></trk></gpx>`,
	} {
		t.Run(name, func(t *testing.T) {
			res, err := im.ImportGPX(context.Background(), "user", "dir/run.gpx", []byte(doc), Options{})
			if err != nil {
				t.Fatalf("ImportGPX returned error %v; invalid files must be reported in the Result", err)
			}
			if res.Status != StatusFailed || res.Error == "" || res.File != "dir/run.gpx" {
				t.Errorf("got %+v, want status %q with an error message", res, StatusFailed)
			}
		})
	}
}

func TestImportGPXReportsInvalidPositionLocation(t *testing.T) {
	doc := `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg>
		<trkpt lat="59.3" lon="18.0"/><trkpt lat="95" lon="18.1"/>
	</trkseg></trk></gpx>`
	res, err := New(nil).ImportGPX(context.Background(), "user", "run.gpx", []byte(doc), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Error, "point 2 of track 1, segment 1 has latitude 95") {
		t.Errorf("Error = %q, want it to say which point is wrong", res.Error)
	}
}
