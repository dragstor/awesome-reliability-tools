package favicons

import (
	"context"
	"time"
)

// Target is the minimal information needed to fetch a favicon.
type Target struct {
	ID       string
	Homepage string
}

// Fetcher fetches a favicon for a homepage, returning the extension and bytes.
// It is a function value so tests can substitute a deterministic fake.
type Fetcher func(ctx context.Context, homepage string) (ext string, data []byte, err error)

// Summary reports the outcome of a favicon run.
type Summary struct {
	Fetched  int
	Skipped  int
	Failed   int
	Failures []string
}

// Logf receives one human-readable line per tool. It may be nil.
type Logf func(format string, args ...any)

// Run fetches favicons for targets into dir.
//
// When force is false, tools that already have a cached favicon are skipped.
// When force is true, they are re-fetched, but an existing file is only ever
// replaced after a successful download, so a temporary remote failure cannot
// destroy a valid cached icon. One failing tool never aborts the run.
func Run(ctx context.Context, targets []Target, dir string, force bool, fetch Fetcher, logf Logf) Summary {
	if logf == nil {
		logf = func(string, ...any) {}
	}
	var summary Summary
	for _, t := range targets {
		if t.Homepage == "" {
			continue
		}
		if !force {
			if name, ok := Find(dir, t.ID); ok {
				summary.Skipped++
				logf("skip   %s (%s)", t.ID, name)
				continue
			}
		}

		reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		ext, data, err := fetch(reqCtx, t.Homepage)
		cancel()
		if err != nil {
			summary.Failed++
			summary.Failures = append(summary.Failures, t.ID+": "+err.Error())
			logf("fail   %s: %v", t.ID, err)
			continue
		}
		if err := Save(dir, t.ID, ext, data); err != nil {
			summary.Failed++
			summary.Failures = append(summary.Failures, t.ID+": "+err.Error())
			logf("fail   %s: %v", t.ID, err)
			continue
		}
		summary.Fetched++
		logf("fetch  %s (%s, %d bytes)", t.ID, ext, len(data))
	}
	return summary
}
