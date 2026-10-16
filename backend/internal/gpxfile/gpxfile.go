// Package gpxfile turns GPX files into activities ready to be stored.
package gpxfile

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tkrajina/gpxgo/gpx"
)

// ErrNoTrack is returned when a file has no track or route with at least two points.
var ErrNoTrack = errors.New("gpx file contains no track with at least two points")

// Activity is the parsed content of one GPX file. All tracks (or, if the
// file has none, routes) in the file become lines of the same activity.
type Activity struct {
	Name      string
	Sport     string
	StartedAt *time.Time
	DurationS *int
	// Lines holds one []LonLat per track segment.
	Lines [][]LonLat
}

// LonLat is a WGS 84 coordinate.
type LonLat struct {
	Lon, Lat float64
}

// Parse parses a GPX document. fallbackName is used when the file has no name.
func Parse(data []byte, fallbackName string) (Activity, error) {
	g, err := gpx.ParseBytes(data)
	if err != nil {
		return Activity{}, fmt.Errorf("parse gpx: %w", err)
	}

	a := Activity{Name: g.Name}
	for _, trk := range g.Tracks {
		if a.Name == "" {
			a.Name = trk.Name
		}
		if a.Sport == "" {
			a.Sport = trk.Type
		}
		for _, seg := range trk.Segments {
			a.addLine(seg.Points)
		}
	}
	if len(a.Lines) == 0 {
		for _, rte := range g.Routes {
			if a.Name == "" {
				a.Name = rte.Name
			}
			a.addLine(rte.Points)
		}
	}
	if len(a.Lines) == 0 {
		return Activity{}, ErrNoTrack
	}

	if a.Name == "" {
		a.Name = fallbackName
	}
	a.Sport = normalizeSport(a.Sport)

	if tb := g.TimeBounds(); !tb.StartTime.IsZero() {
		start := tb.StartTime.UTC()
		a.StartedAt = &start
		if !tb.EndTime.IsZero() {
			d := int(tb.EndTime.Sub(tb.StartTime).Seconds())
			a.DurationS = &d
		}
	} else if g.Time != nil {
		start := g.Time.UTC()
		a.StartedAt = &start
	}

	return a, nil
}

func (a *Activity) addLine(points []gpx.GPXPoint) {
	if len(points) < 2 {
		return
	}
	line := make([]LonLat, 0, len(points))
	for _, p := range points {
		line = append(line, LonLat{Lon: p.Longitude, Lat: p.Latitude})
	}
	a.Lines = append(a.Lines, line)
}

// WKT returns the activity's lines as a WKT MULTILINESTRING in EPSG:4326.
func (a Activity) WKT() string {
	var b strings.Builder
	b.WriteString("MULTILINESTRING(")
	for i, line := range a.Lines {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('(')
		for j, p := range line {
			if j > 0 {
				b.WriteByte(',')
			}
			b.WriteString(strconv.FormatFloat(p.Lon, 'f', -1, 64))
			b.WriteByte(' ')
			b.WriteString(strconv.FormatFloat(p.Lat, 'f', -1, 64))
		}
		b.WriteByte(')')
	}
	b.WriteByte(')')
	return b.String()
}

// sportAliases maps the activity types written by common exporters to one
// name per sport. Strava's older exports use numeric codes.
var sportAliases = map[string]string{
	"1":               "cycling",
	"9":               "running",
	"10":              "walking",
	"4":               "hiking",
	"run":             "running",
	"trail_running":   "running",
	"ride":            "cycling",
	"biking":          "cycling",
	"road_biking":     "cycling",
	"mountain_biking": "cycling",
	"walk":            "walking",
	"hike":            "hiking",
	"swim":            "swimming",
}

func normalizeSport(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "unknown"
	}
	if alias, ok := sportAliases[s]; ok {
		return alias
	}
	return s
}
