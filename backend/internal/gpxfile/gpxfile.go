// Package gpxfile turns GPX files into activities ready to be stored.
package gpxfile

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/tkrajina/gpxgo/gpx"

	"github.com/jlindbergh/my-tracks/backend/internal/activity"
)

// ErrNoTrack is returned when a file has no track or route with at least
// two GPS points, e.g. an indoor or treadmill activity. The returned error is
// a *NoTrackError; use errors.As to get the activity's name and sport.
var ErrNoTrack = errors.New("gpx file contains no GPS track")

// NoTrackError describes a file that was readable but has no GPS track.
type NoTrackError struct {
	// Name is the activity name written in the file, or "" if it has none.
	// The fallback name is not used, since nothing is stored.
	Name  string
	Sport string
}

func (e *NoTrackError) Error() string { return ErrNoTrack.Error() }

// Is makes errors.Is(err, ErrNoTrack) true.
func (e *NoTrackError) Is(target error) bool { return target == ErrNoTrack }

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
	a.Sport = normalizeSport(a.Sport)

	if len(a.Lines) == 0 {
		return Activity{}, &NoTrackError{Name: a.Name, Sport: a.Sport}
	}
	if a.Name == "" {
		a.Name = fallbackName
	}

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
	line := make([]LonLat, 0, len(points))
	for _, p := range points {
		// Some exporters write 0,0 for points without a GPS fix.
		if p.Latitude == 0 && p.Longitude == 0 {
			continue
		}
		line = append(line, LonLat{Lon: p.Longitude, Lat: p.Latitude})
	}
	if len(line) < 2 {
		return
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

// sportAliases maps the activity types written by common exporters to the
// sport ids in package activity. Strava's older exports use numeric codes.
var sportAliases = map[string]string{
	"1":               activity.SportCycling,
	"9":               activity.SportRunning,
	"10":              activity.SportWalking,
	"4":               activity.SportHiking,
	"run":             activity.SportRunning,
	"trail_running":   activity.SportRunning,
	"ride":            activity.SportCycling,
	"biking":          activity.SportCycling,
	"road_biking":     activity.SportCycling,
	"mountain_biking": activity.SportCycling,
	"walk":            activity.SportWalking,
	"hike":            activity.SportHiking,
	"swim":            activity.SportSwimming,
}

// normalizeSport lowercases a sport type, turning CamelCase ("VirtualRide",
// as Strava writes it) and spaces into snake_case, then applies sportAliases.
func normalizeSport(s string) string {
	var b strings.Builder
	prevLower := false
	for _, r := range strings.TrimSpace(s) {
		switch {
		case r == ' ' || r == '-':
			r = '_'
		case unicode.IsUpper(r) && prevLower:
			b.WriteByte('_')
		}
		prevLower = unicode.IsLower(r)
		b.WriteRune(unicode.ToLower(r))
	}
	s = b.String()
	if s == "" {
		return activity.SportUnknown
	}
	if alias, ok := sportAliases[s]; ok {
		return alias
	}
	return s
}
