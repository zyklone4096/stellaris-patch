package main

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
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
