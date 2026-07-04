package main

import (
	"context"
	"errors"
	"os"

	"github.com/urfave/cli/v3"
	"github.com/zyklone4096/stellaris-patch/patcher"
)

func main() {
	if err := (&cli.Command{
		Name: "stellaris-patch",
		Commands: []*cli.Command{
			{
				Name:        "apply-patch",
				Action:      applyPatch,
				Category:    "patching",
				Description: "Apply patches in current workspace",
				Arguments: []cli.Argument{
					&cli.StringArgs{
						Name: "files",
						Min:  0,
						Max:  -1,
					},
				},
			},
			{
				Name:        "rebuild-patch",
				Action:      rebuildPatch,
				Category:    "patching",
				Description: "Rebuild patches for current workspace",
				Arguments: []cli.Argument{
					&cli.StringArgs{
						Name: "files",
						Min:  0,
						Max:  -1,
					},
				},
			},
			{
				Name:        "init-workspace",
				Action:      initWorkspace,
				Category:    "workspace",
				Description: "Initialize new stellaris-patch mod workspace in current directory",
			},
			{
				Name:        "deploy",
				Action:      deploy,
				Category:    "workspace",
				Description: "Generate and deploy mod",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "target",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "name",
						Value: "stellaris_patch",
						Usage: "Target mod name",
					},
					&cli.BoolFlag{
						Name:  "purge",
						Value: false,
						Usage: "Purge existing mod only (won't deploy)",
					},
				},
			},
			{
				Name:        "install",
				Action:      hardInstall,
				Category:    "workspace",
				Description: "Install files to game directory, this should not be used unless your mod requires",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "purge",
						Value: false,
						Usage: "Recover original game files (won't install), recommended before base game update",
					},
					&cli.BoolFlag{
						Name:  "force",
						Value: false,
						Usage: "Force mode for purge, won't check",
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "base",
				Value: "",
				Usage: "Base game directory",
			},
		},
	}).Run(context.Background(), os.Args); err != nil {
		panic(err)
	}
}

type current struct {
	workspace string
	game      string
}

func getCurrent(cmd *cli.Command) (current, error) {
	result := current{}

	wd, err := os.Getwd()
	if err != nil {
		return result, err
	}

	base := cmd.String("base")
	if base == "" {
		base = os.Getenv("STELLARIS_HOME")
	}
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return result, errors.New("base game not found")
	}

	result.workspace = wd
	result.game = base
	return result, nil
}

func (c current) NewPatcher() (*patcher.Patcher, error) {
	return patcher.NewPatcher(c.workspace, c.game)
}
