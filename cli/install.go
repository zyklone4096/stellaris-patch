package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

func hardInstall(_ context.Context, cmd *cli.Command) error {
	c, err := getCurrent(cmd)
	if err != nil {
		return err
	}
	p, err := c.NewPatcher()
	if err != nil {
		return err
	}

	if cmd.Bool("purge") {
		return p.HardPurge(cmd.Bool("hard"))
	}

	return p.HardInstall()
}
