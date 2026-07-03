package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

func initWorkspace(_ context.Context, _ *cli.Command) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err = os.WriteFile(filepath.Join(wd, ".gitignore"), []byte(`
src
metadata
`), 0644); err != nil {
		return err
	}

	if err = os.Mkdir(filepath.Join(wd, "src"), 0755); err != nil {
		return err
	}
	if err = os.Mkdir(filepath.Join(wd, "patches"), 0755); err != nil {
		return err
	}
	if err = os.Mkdir(filepath.Join(wd, "metadata"), 0755); err != nil {
		return err
	}

	fmt.Printf("Initialized new workspace in %s\n", wd)

	return nil
}
