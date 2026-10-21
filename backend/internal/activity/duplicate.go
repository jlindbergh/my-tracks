package activity

import (
	"math"
	"time"
)

// DuplicateRule decides when an activity is a possible duplicate of another
// one by the same user: start times within StartTolerance, and distance and
// duration within RelativeTolerance (a fraction) of the larger value.
// Duration is skipped when either one is unknown.
type DuplicateRule struct {
	StartTolerance    time.Duration
	RelativeTolerance float64
}

// Comparison describes how two activities differ.
type Comparison struct {
	// StartDiffS is how far apart the start times are, in seconds, or nil
	// if either start time is unknown.
	StartDiffS *float64 `json:"startDiffS"`
	// DistanceDiff is the difference in distance, as a fraction of the longer one.
	DistanceDiff float64 `json:"distanceDiff"`
	// DurationDiff is the difference in duration, as a fraction of the longer
	// one, or nil if either duration is unknown.
	DurationDiff *float64 `json:"durationDiff"`
}

// Compare compares the activities a and b.
func Compare(a, b Facts) Comparison {
	c := Comparison{DistanceDiff: relativeDiff(a.DistanceM, b.DistanceM)}
	if a.StartedAt != nil && b.StartedAt != nil {
		s := math.Abs(a.StartedAt.Sub(*b.StartedAt).Seconds())
		c.StartDiffS = &s
	}
	if a.DurationS != nil && b.DurationS != nil {
		d := relativeDiff(float64(*a.DurationS), float64(*b.DurationS))
		c.DurationDiff = &d
	}
	return c
}

// Matches reports whether activities that compare as c are possible duplicates.
// Activities without known start times never are.
func (r DuplicateRule) Matches(c Comparison) bool {
	return c.StartDiffS != nil && *c.StartDiffS <= r.StartTolerance.Seconds() &&
		c.DistanceDiff <= r.RelativeTolerance &&
		(c.DurationDiff == nil || *c.DurationDiff <= r.RelativeTolerance)
}

// relativeDiff returns the difference between a and b as a fraction of the larger.
func relativeDiff(a, b float64) float64 {
	larger := math.Max(a, b)
	if larger == 0 {
		return 0
	}
	return math.Abs(a-b) / larger
}
