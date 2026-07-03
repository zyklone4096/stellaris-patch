package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func applyPatch(_ context.Context, cmd *cli.Command) error {
	c, err := getCurrent(cmd)
	if err != nil {
		return err
	}

	base := cmd.String("base")
	if base == "" {
		base = os.Getenv("STELLARIS_HOME")
	}

	if _, err := os.Stat(base); os.IsNotExist(err) {
		return errors.New("base game not found")
	}

	if p, err := c.NewPatcher(); err == nil {
		targets := cmd.StringArgs("files")
		all := len(targets)
		if all == 0 {
			fmt.Println("Applying all patches")
			if err = p.ApplyAll(); err != nil {
				return err
			}
		}

		for i, target := range targets {
			fmt.Printf("Applying %s (%d/%d)", target, i, all)
			if err = p.Apply(target); err != nil {
				fmt.Println("Error applying ", target)
				return err
			}
		}
	} else {
		return err
	}

	return nil
}

func rebuildPatch(_ context.Context, cmd *cli.Command) error {
	c, err := getCurrent(cmd)
	if err != nil {
		return err
	}

	base := cmd.String("base")
	if base == "" {
		base = os.Getenv("STELLARIS_HOME")
	}
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return errors.New("base game not found")
	}

	if p, err := c.NewPatcher(); err == nil {
		targets := cmd.StringArgs("files")
		all := len(targets)
		if all == 0 {
			fmt.Println("Regenerating for all changed files")
			return p.RegenerateChanged()
		}

		for idx, file := range targets {
			fmt.Printf("Regenerating %s (%d/%d)", file, idx, all)
			if err = p.Generate(file); err != nil {
				fmt.Println("Error regenerating ", file)
				return err
			}
		}
	} else {
		return err
	}

	return nil
}
