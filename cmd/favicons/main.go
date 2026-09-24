package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/dragstor/awesome-reliability-tools/internal/data"
	"github.com/dragstor/awesome-reliability-tools/internal/favicons"
	"github.com/dragstor/awesome-reliability-tools/internal/favicons/fetch"
)

func main() {
	force := flag.Bool("force", false, "re-fetch favicons that are already cached")
	dir := flag.String("dir", "site/assets/favicons", "directory to store cached favicons")
	flag.Parse()

	tools, err := data.LoadTools("data/tools.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	targets := make([]favicons.Target, 0, len(tools))
	for _, t := range tools {
		targets = append(targets, favicons.Target{ID: t.ID, Homepage: t.Homepage})
	}

	logf := func(format string, args ...any) {
		fmt.Printf(format+"\n", args...)
	}

	summary := favicons.Run(context.Background(), targets, *dir, *force, fetch.Fetch, logf)

	fmt.Printf("\nfavicons: fetched=%d skipped=%d failed=%d\n", summary.Fetched, summary.Skipped, summary.Failed)
	if len(summary.Failures) > 0 {
		fmt.Println("failures:")
		for _, f := range summary.Failures {
			fmt.Printf("  - %s\n", f)
		}
	}
}
