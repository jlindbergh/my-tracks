package gpxfile

import (
	"errors"
	"testing"
	"time"
)

const twoSegments = `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" creator="test" xmlns="http://www.topografix.com/GPX/1/1">
  <trk>
    <name>Morning run</name>
    <type>9</type>
    <trkseg>
      <trkpt lat="59.3293" lon="18.0686"><time>2026-05-01T06:00:00Z</time></trkpt>
      <trkpt lat="59.3300" lon="18.0700"><time>2026-05-01T06:01:00Z</time></trkpt>
    </trkseg>
    <trkseg>
      <trkpt lat="59.3310" lon="18.0710"><time>2026-05-01T06:05:00Z</time></trkpt>
      <trkpt lat="59.3320" lon="18.0720"><time>2026-05-01T06:30:00Z</time></trkpt>
    </trkseg>
  </trk>
</gpx>`

func TestParse(t *testing.T) {
	a, err := Parse([]byte(twoSegments), "fallback")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if a.Name != "Morning run" {
		t.Errorf("Name = %q, want %q", a.Name, "Morning run")
	}
	if a.Sport != "running" {
		t.Errorf("Sport = %q, want %q", a.Sport, "running")
	}
	if want := time.Date(2026, 5, 1, 6, 0, 0, 0, time.UTC); a.StartedAt == nil || !a.StartedAt.Equal(want) {
		t.Errorf("StartedAt = %v, want %v", a.StartedAt, want)
	}
	if a.DurationS == nil || *a.DurationS != 30*60 {
		t.Errorf("DurationS = %v, want %d", a.DurationS, 30*60)
	}
	want := "MULTILINESTRING((18.0686 59.3293,18.07 59.33),(18.071 59.331,18.072 59.332))"
	if got := a.WKT(); got != want {
		t.Errorf("WKT() = %q, want %q", got, want)
	}
}

func TestParseFallbackNameAndUnknownSport(t *testing.T) {
	doc := `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg>
		<trkpt lat="1" lon="2"/><trkpt lat="3" lon="4"/>
	</trkseg></trk></gpx>`
	a, err := Parse([]byte(doc), "my-file")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if a.Name != "my-file" || a.Sport != "unknown" || a.StartedAt != nil {
		t.Errorf("got Name=%q Sport=%q StartedAt=%v", a.Name, a.Sport, a.StartedAt)
	}
}

func TestParseNoTrack(t *testing.T) {
	doc := `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg>
		<trkpt lat="1" lon="2"/>
	</trkseg></trk></gpx>`
	_, err := Parse([]byte(doc), "x")
	if !errors.Is(err, ErrNoTrack) {
		t.Fatalf("err = %v, want ErrNoTrack", err)
	}
	var nt *NoTrackError
	if errors.As(err, &nt) && nt.Name != "" {
		t.Errorf("Name = %q, want \"\": the fallback name is not used without a track", nt.Name)
	}
}

func TestParseNoTrackReportsActivity(t *testing.T) {
	tests := map[string]string{
		"no segments": `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1">
			<trk><name>Zwift ride</name><type>VirtualRide</type></trk></gpx>`,
		"only 0,0 points": `<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1">
			<trk><name>Zwift ride</name><type>VirtualRide</type><trkseg>
			<trkpt lat="0" lon="0"/><trkpt lat="0" lon="0"/><trkpt lat="0" lon="0"/>
			</trkseg></trk></gpx>`,
	}
	for name, doc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := Parse([]byte(doc), "x")
			var nt *NoTrackError
			if !errors.As(err, &nt) {
				t.Fatalf("err = %v, want *NoTrackError", err)
			}
			if nt.Name != "Zwift ride" || nt.Sport != "virtual_ride" {
				t.Errorf("got Name=%q Sport=%q", nt.Name, nt.Sport)
			}
		})
	}
}

func TestNormalizeSport(t *testing.T) {
	for in, want := range map[string]string{
		"":               "unknown",
		" Running ":      "running",
		"Run":            "running",
		"VirtualRide":    "virtual_ride",
		"indoor_cycling": "indoor_cycling",
		"Indoor Rowing":  "indoor_rowing",
		"9":              "running",
	} {
		if got := normalizeSport(in); got != want {
			t.Errorf("normalizeSport(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	if _, err := Parse([]byte("not xml"), "x"); err == nil {
		t.Error("expected an error for invalid input")
	}
}
