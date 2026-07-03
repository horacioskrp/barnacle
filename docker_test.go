package main

import (
	"testing"
	"time"

	"github.com/moby/moby/api/types/build"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/api/types/volume"
)

func TestIsDangling(t *testing.T) {
	tests := []struct {
		name string
		img  image.Summary
		want bool
	}{
		{"no tags at all", image.Summary{RepoTags: nil}, true},
		{"single none tag", image.Summary{RepoTags: []string{"<none>:<none>"}}, true},
		{"real tag", image.Summary{RepoTags: []string{"myimage:latest"}}, false},
		{"mixed tags", image.Summary{RepoTags: []string{"myimage:latest", "<none>:<none>"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDangling(tt.img); got != tt.want {
				t.Errorf("isDangling() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsStoppedState(t *testing.T) {
	tests := []struct {
		state container.ContainerState
		want  bool
	}{
		{container.StateExited, true},
		{container.StateCreated, true},
		{container.StateDead, true},
		{container.StateRunning, false},
		{container.StatePaused, false},
		{container.StateRestarting, false},
		{container.StateRemoving, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := isStoppedState(tt.state); got != tt.want {
				t.Errorf("isStoppedState(%s) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

func TestAnalyzeDanglingImages(t *testing.T) {
	old := time.Now().Add(-10 * 24 * time.Hour)
	recent := time.Now().Add(-1 * time.Hour)

	items := []image.Summary{
		{RepoTags: nil, Size: 100, Created: old.Unix()},
		{RepoTags: []string{"<none>:<none>"}, Size: 200, Created: recent.Unix()},
		{RepoTags: []string{"kept:latest"}, Size: 999, Created: recent.Unix()},
	}

	cat := analyzeDanglingImages(items)

	if cat.ID != CategoryDanglingImages {
		t.Errorf("ID = %v, want CategoryDanglingImages", cat.ID)
	}
	if cat.Count != 2 {
		t.Errorf("Count = %d, want 2", cat.Count)
	}
	if cat.Size != 300 {
		t.Errorf("Size = %d, want 300", cat.Size)
	}
	if !cat.HasStale {
		t.Errorf("HasStale = false, want true (oldest item is 10 days old)")
	}
}

func TestAnalyzeStoppedContainers(t *testing.T) {
	recent := time.Now().Add(-1 * time.Hour)

	items := []container.Summary{
		{State: container.StateExited, SizeRw: 50, Created: recent.Unix()},
		{State: container.StateRunning, SizeRw: 999, Created: recent.Unix()},
		{State: container.StateDead, SizeRw: 25, Created: recent.Unix()},
	}

	cat := analyzeStoppedContainers(items)

	if cat.Count != 2 {
		t.Errorf("Count = %d, want 2", cat.Count)
	}
	if cat.Size != 75 {
		t.Errorf("Size = %d, want 75", cat.Size)
	}
	if cat.HasStale {
		t.Errorf("HasStale = true, want false (items are recent)")
	}
}

func TestAnalyzeOrphanVolumes(t *testing.T) {
	old := time.Now().Add(-8 * 24 * time.Hour).Format(time.RFC3339)
	recent := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)

	items := []volume.Volume{
		{CreatedAt: old, UsageData: &volume.UsageData{RefCount: 0, Size: 10}},
		{CreatedAt: recent, UsageData: &volume.UsageData{RefCount: 1, Size: 20}},
		{CreatedAt: recent, UsageData: nil},
	}

	cat := analyzeOrphanVolumes(items)

	if cat.Count != 2 {
		t.Errorf("Count = %d, want 2 (RefCount=0 and nil usage data)", cat.Count)
	}
	if cat.Size != 10 {
		t.Errorf("Size = %d, want 10", cat.Size)
	}
	if !cat.HasStale {
		t.Errorf("HasStale = false, want true (oldest volume is 8 days old)")
	}
}

func TestAnalyzeBuildCache(t *testing.T) {
	old := time.Now().Add(-30 * 24 * time.Hour)
	recent := time.Now().Add(-1 * time.Hour)

	items := []build.CacheRecord{
		{InUse: false, Size: 111, CreatedAt: old},
		{InUse: true, Size: 999, CreatedAt: recent},
		{InUse: false, Size: 222, CreatedAt: recent},
	}

	cat := analyzeBuildCache(items)

	if cat.Count != 2 {
		t.Errorf("Count = %d, want 2", cat.Count)
	}
	if cat.Size != 333 {
		t.Errorf("Size = %d, want 333", cat.Size)
	}
	if !cat.HasStale {
		t.Errorf("HasStale = false, want true (oldest cache record is 30 days old)")
	}
}

func TestFinalizeAgeZeroTime(t *testing.T) {
	cat := Category{}
	finalizeAge(&cat, time.Time{})

	if cat.HasStale {
		t.Errorf("HasStale = true, want false when no oldest time was recorded")
	}
	if cat.OldestAge != 0 {
		t.Errorf("OldestAge = %v, want 0", cat.OldestAge)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		size int64
		want string
	}{
		{-42, "0 B"},
		{0, "0 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{4404019, "4.2 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := FormatBytes(tt.size); got != tt.want {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.size, got, tt.want)
			}
		})
	}
}

func TestPruneSummaryTotalReclaimed(t *testing.T) {
	summary := PruneSummary{
		Results: []PruneResult{
			{Label: "a", SpaceReclaimed: 100},
			{Label: "b", SpaceReclaimed: 500, Err: errBoom},
			{Label: "c", SpaceReclaimed: 50},
		},
	}

	if got := summary.TotalReclaimed(); got != 150 {
		t.Errorf("TotalReclaimed() = %d, want 150 (the failed result must not count)", got)
	}
}

var errBoom = &testError{"boom"}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
