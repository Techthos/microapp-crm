// Command microapp-crm is a local-first, single-user sales CRM.
//
// It runs as one self-contained binary backed by a single embedded bbolt file,
// and exposes the same data through two surfaces selected at launch:
//
//	microapp-crm -mode tui   # interactive terminal UI (default)
//	microapp-crm -mode mcp   # MCP stdio server for an AI assistant
//
// The two modes never run at once — bbolt holds a process-wide write lock.
// See docs/SPECIFICATIONS.md for the full contract.
package main

import (
	"flag"
	"fmt"
	"os"
)

// version is stamped at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "microapp-crm:", err)
		os.Exit(1)
	}
}

// run parses flags and dispatches to the selected surface. It returns an error
// instead of exiting so it stays testable.
func run(args []string) error {
	fs := flag.NewFlagSet("microapp-crm", flag.ContinueOnError)
	mode := fs.String("mode", "tui", "surface to start: tui | mcp")
	showVersion := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *showVersion {
		fmt.Println("microapp-crm", version)
		return nil
	}

	switch *mode {
	case "tui", "mcp":
		// TODO: wire internal/db, then start the chosen surface
		// (internal/tui or internal/server). Not yet implemented.
		return fmt.Errorf("mode %q not yet implemented", *mode)
	default:
		return fmt.Errorf("unknown mode %q (want tui or mcp)", *mode)
	}
}
