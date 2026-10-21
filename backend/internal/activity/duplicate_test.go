package activity

import (
	"testing"
	"time"
)

func facts(start time.Time, durationS int, distanceM float64) Facts {
	return Facts{StartedAt: &start, DurationS: &durationS, DistanceM: distanceM}
}

func TestCompare(t *testing.T) {
	start := time.Date(2026, 6, 1, 7, 0, 0, 0, time.UTC)
	c := Compare(facts(start, 3600, 10000), facts(start.Add(-1500*time.Millisecond), 3528, 10100))

	if c.StartDiffS == nil || *c.StartDiffS != 1.5 {
		t.Errorf("StartDiffS = %v, want 1.5", c.StartDiffS)
	}
	if want := 100.0 / 10100; c.DistanceDiff != want {
		t.Errorf("DistanceDiff = %v, want %v", c.DistanceDiff, want)
	}
	if c.DurationDiff == nil || *c.DurationDiff != 0.02 {
		t.Errorf("DurationDiff = %v, want 0.02", c.DurationDiff)
	}
}

func TestCompareUnknownValues(t *testing.T) {
	c := Compare(Facts{}, Facts{})
	if c.StartDiffS != nil || c.DurationDiff != nil || c.DistanceDiff != 0 {
		t.Errorf("got %+v, want only a zero DistanceDiff", c)
	}
}

func TestDuplicateRuleMatches(t *testing.T) {
	rule := DuplicateRule{StartTolerance: 2 * time.Second, RelativeTolerance: 0.02}
	start := time.Date(2026, 6, 1, 7, 0, 0, 0, time.UTC)
	base := facts(start, 3600, 10000)
	noDuration := base
	noDuration.DurationS = nil
	noStart := base
	noStart.StartedAt = nil

	tests := []struct {
		name  string
		other Facts
		want  bool
	}{
		{"identical", base, true},
		{"start 2s apart", facts(start.Add(2*time.Second), 3600, 10000), true},
		{"start 3s apart", facts(start.Add(3*time.Second), 3600, 10000), false},
		{"distance 2% longer", facts(start, 3600, 10200), true},
		{"distance 3% longer", facts(start, 3600, 10300), false},
		{"duration 3% longer", facts(start, 3708, 10000), false},
		{"duration unknown", noDuration, true},
		{"start unknown", noStart, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rule.Matches(Compare(base, tt.other)); got != tt.want {
				t.Errorf("Matches = %v, want %v", got, tt.want)
			}
		})
	}
}
