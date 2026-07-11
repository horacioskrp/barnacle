// Package docker implements Barnacle's Docker Engine interaction: analyzing
// reclaimable disk usage and running targeted prune operations.
package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/moby/moby/api/types/build"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
)

// staleThreshold is the age above which an unused resource is flagged as
// "stale" in the UI (the smart alert on old, forgotten resources).
const staleThreshold = 7 * 24 * time.Hour

// CategoryID identifies one of the four classes of reclaimable Docker
// resources that Barnacle knows how to analyze and clean.
type CategoryID int

const (
	CategoryDanglingImages CategoryID = iota
	CategoryStoppedContainers
	CategoryOrphanVolumes
	CategoryBuildCache
)

// Category holds the disk-usage analysis for one class of resources.
type Category struct {
	ID        CategoryID
	Label     string
	Count     int
	Size      int64
	OldestAge time.Duration
	HasStale  bool
}

// PruneResult is the outcome of cleaning a single category.
type PruneResult struct {
	Label          string
	SpaceReclaimed uint64
	Err            error
}

// PruneSummary aggregates the outcome of a full cleanup run.
type PruneSummary struct {
	Results []PruneResult
}

// TotalReclaimed sums the space freed across every successful result.
func (s PruneSummary) TotalReclaimed() uint64 {
	var total uint64
	for _, r := range s.Results {
		if r.Err == nil {
			total += r.SpaceReclaimed
		}
	}
	return total
}

// Client wraps the Docker Engine API client used by Barnacle.
type Client struct {
	cli *client.Client
}

// NewClient connects to the Docker daemon using the standard environment
// configuration (DOCKER_HOST, DOCKER_CERT_PATH, TLS...), which defaults to
// the local unix socket /var/run/docker.sock.
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("connexion au démon Docker impossible: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := cli.Ping(ctx, client.PingOptions{}); err != nil {
		return nil, fmt.Errorf("le démon Docker ne répond pas (socket monté ? droits suffisants ?): %w", err)
	}

	return &Client{cli: cli}, nil
}

// Close releases the underlying HTTP client resources.
func (d *Client) Close() error {
	return d.cli.Close()
}

// Analyze fetches disk usage from the daemon and buckets it into the four
// categories Barnacle knows how to clean, each annotated with a staleness
// flag when its oldest unused item is older than staleThreshold.
func (d *Client) Analyze(ctx context.Context) ([]Category, error) {
	usage, err := d.cli.DiskUsage(ctx, client.DiskUsageOptions{
		Containers: true,
		Images:     true,
		Volumes:    true,
		BuildCache: true,
		Verbose:    true,
	})
	if err != nil {
		return nil, fmt.Errorf("récupération de l'espace disque Docker: %w", err)
	}

	return []Category{
		analyzeDanglingImages(usage.Images.Items),
		analyzeStoppedContainers(usage.Containers.Items),
		analyzeOrphanVolumes(usage.Volumes.Items),
		analyzeBuildCache(usage.BuildCache.Items),
	}, nil
}

func analyzeDanglingImages(items []image.Summary) Category {
	cat := Category{ID: CategoryDanglingImages, Label: "Images suspendues"}
	var oldest time.Time

	for _, img := range items {
		if !isDangling(img) {
			continue
		}
		cat.Count++
		cat.Size += img.Size
		created := time.Unix(img.Created, 0)
		if oldest.IsZero() || created.Before(oldest) {
			oldest = created
		}
	}

	finalizeAge(&cat, oldest)
	return cat
}

// isDangling reports whether an image has no repository tag, i.e. is an
// orphaned "<none>:<none>" image left behind by a rebuild.
func isDangling(img image.Summary) bool {
	if len(img.RepoTags) == 0 {
		return true
	}
	for _, tag := range img.RepoTags {
		if tag != "<none>:<none>" {
			return false
		}
	}
	return true
}

func analyzeStoppedContainers(items []container.Summary) Category {
	cat := Category{ID: CategoryStoppedContainers, Label: "Conteneurs arrêtés"}
	var oldest time.Time

	for _, c := range items {
		if !isStoppedState(c.State) {
			continue
		}
		cat.Count++
		cat.Size += c.SizeRw
		created := time.Unix(c.Created, 0)
		if oldest.IsZero() || created.Before(oldest) {
			oldest = created
		}
	}

	finalizeAge(&cat, oldest)
	return cat
}

func isStoppedState(state container.ContainerState) bool {
	switch state {
	case container.StateExited, container.StateCreated, container.StateDead:
		return true
	default:
		return false
	}
}

func analyzeOrphanVolumes(items []volume.Volume) Category {
	cat := Category{ID: CategoryOrphanVolumes, Label: "Volumes orphelins"}
	var oldest time.Time

	for _, v := range items {
		if v.UsageData != nil && v.UsageData.RefCount > 0 {
			continue
		}
		cat.Count++
		if v.UsageData != nil {
			cat.Size += v.UsageData.Size
		}
		if created, err := time.Parse(time.RFC3339, v.CreatedAt); err == nil {
			if oldest.IsZero() || created.Before(oldest) {
				oldest = created
			}
		}
	}

	finalizeAge(&cat, oldest)
	return cat
}

func analyzeBuildCache(items []build.CacheRecord) Category {
	cat := Category{ID: CategoryBuildCache, Label: "Cache de build obsolète"}
	var oldest time.Time

	for _, b := range items {
		if b.InUse {
			continue
		}
		cat.Count++
		cat.Size += b.Size
		if oldest.IsZero() || b.CreatedAt.Before(oldest) {
			oldest = b.CreatedAt
		}
	}

	finalizeAge(&cat, oldest)
	return cat
}

func finalizeAge(cat *Category, oldest time.Time) {
	if oldest.IsZero() {
		return
	}
	cat.OldestAge = time.Since(oldest)
	cat.HasStale = cat.OldestAge > staleThreshold
}

// Prune removes the resources for every category whose ID maps to true in
// selected. It keeps going even if one step fails, collecting every error
// into the returned summary instead of aborting the whole run.
func (d *Client) Prune(ctx context.Context, selected map[CategoryID]bool) PruneSummary {
	var summary PruneSummary

	if selected[CategoryStoppedContainers] {
		summary.Results = append(summary.Results, d.pruneContainers(ctx))
	}
	if selected[CategoryDanglingImages] {
		summary.Results = append(summary.Results, d.pruneImages(ctx))
	}
	if selected[CategoryOrphanVolumes] {
		summary.Results = append(summary.Results, d.pruneVolumes(ctx))
	}
	if selected[CategoryBuildCache] {
		summary.Results = append(summary.Results, d.pruneBuildCache(ctx))
	}

	return summary
}

func (d *Client) pruneContainers(ctx context.Context) PruneResult {
	res := PruneResult{Label: "Conteneurs arrêtés"}
	report, err := d.cli.ContainerPrune(ctx, client.ContainerPruneOptions{})
	if err != nil {
		res.Err = err
		return res
	}
	res.SpaceReclaimed = report.Report.SpaceReclaimed
	return res
}

func (d *Client) pruneImages(ctx context.Context) PruneResult {
	res := PruneResult{Label: "Images suspendues"}
	filters := client.Filters{}.Add("dangling", "true")
	report, err := d.cli.ImagePrune(ctx, client.ImagePruneOptions{Filters: filters})
	if err != nil {
		res.Err = err
		return res
	}
	res.SpaceReclaimed = report.Report.SpaceReclaimed
	return res
}

func (d *Client) pruneVolumes(ctx context.Context) PruneResult {
	res := PruneResult{Label: "Volumes orphelins"}
	report, err := d.cli.VolumePrune(ctx, client.VolumePruneOptions{All: true})
	if err != nil {
		res.Err = err
		return res
	}
	res.SpaceReclaimed = report.Report.SpaceReclaimed
	return res
}

func (d *Client) pruneBuildCache(ctx context.Context) PruneResult {
	res := PruneResult{Label: "Cache de build obsolète"}
	report, err := d.cli.BuildCachePrune(ctx, client.BuildCachePruneOptions{All: true})
	if err != nil {
		res.Err = err
		return res
	}
	res.SpaceReclaimed = report.Report.SpaceReclaimed
	return res
}

// FormatBytes converts a byte count into a human friendly string, e.g.
// "512 B", "4.2 MB" or "1.3 GB".
func FormatBytes(size int64) string {
	if size < 0 {
		size = 0
	}
	return formatBytesUint64(uint64(size))
}

func formatBytesUint64(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(b)/float64(div), units[exp])
}
