package main

import (
	"fmt"
	"os"

	"github.com/dragstor/awesome-reliability-tools/internal/data"
	"github.com/dragstor/awesome-reliability-tools/internal/validate"
)

func main() {
	catalog, err := data.LoadCatalog("data/features.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	tools, err := data.LoadTools("data/tools.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	res := validate.Tools(tools, catalog)
	if !res.OK() {
		fmt.Fprintln(os.Stderr, "validation failed:")
		fmt.Fprintln(os.Stderr, res.Error())
		os.Exit(1)
	}
	fmt.Printf("ok: %d feature(s), %d tool(s)\n", len(catalog.Features), len(tools))
}
