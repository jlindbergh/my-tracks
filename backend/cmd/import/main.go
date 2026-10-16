// Command import batch-imports all GPX files in a directory for one user.
//
//	docker compose run --rm backend import -dir /samples
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/jlindbergh/my-tracks/backend/internal/importer"
	"github.com/jlindbergh/my-tracks/backend/internal/store"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "import:", err)
		os.Exit(1)
	}
}

func run() error {
	dir := flag.String("dir", "/samples", "directory to search (recursively) for .gpx files")
	email := flag.String("user", os.Getenv("DEV_USER_EMAIL"), "email of the user who will own the activities")
	flag.Parse()

	if *email == "" {
		return errors.New("-user (or DEV_USER_EMAIL) is required")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is not set")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	st, err := store.New(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer st.Close()

	u, err := st.EnsureUserByEmail(ctx, *email)
	if err != nil {
		return err
	}
	im := importer.New(st)

	counts := map[importer.Status]int{}
	err = filepath.WalkDir(*dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.EqualFold(filepath.Ext(path), ".gpx") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		res, err := im.ImportGPX(ctx, u.ID, path, data)
		if err != nil {
			return err
		}
		counts[res.Status]++
		if res.Error != "" {
			fmt.Printf("%-9s %s: %s\n", res.Status, path, res.Error)
		} else {
			fmt.Printf("%-9s %s\n", res.Status, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("\n%s: %d imported, %d duplicates, %d failed\n", u.Email,
		counts[importer.StatusImported], counts[importer.StatusDuplicate], counts[importer.StatusFailed])
	return nil
}
