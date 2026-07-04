package main

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
)

func initWorkspace(_ context.Context, cmd *cli.Command) error {
	c, err := getCurrent(cmd)
	if err != nil {
		return err
	}
	wd := c.workspace

	if err = os.WriteFile(filepath.Join(wd, ".gitignore"), []byte(`src
metadata
backup
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

func deploy(_ context.Context, cmd *cli.Command) error {
	c, err := getCurrent(cmd)
	if err != nil {
		return err
	}

	target := cmd.String("target")
	if target == "" {
		u, err := user.Current()
		if err != nil {
			return err
		}
		target = filepath.Join(u.HomeDir, "Documents\\Paradox Interactive\\Stellaris\\mod")
	}
	if stats, err := os.Stat(target); os.IsNotExist(err) {
		return err
	} else if err == nil && !stats.IsDir() {
		return fmt.Errorf("%s is not a directory", target)
	}

	name := cmd.String("name")
	descA := filepath.Join(target, name+".mod")
	target = filepath.Join(target, name)

	fmt.Println("Purging mod " + name)
	if err = os.Remove(target); err != nil {
		fmt.Printf("Failed to remove mod directory: %v\n", err)
	}
	if err = os.Remove(descA); err != nil {
		fmt.Printf("Failed to remove mod descriptor: %v\n", err)
	}

	if cmd.Bool("purge") {
		return nil
	}

	fmt.Println("Deploying mod")
	src := filepath.Join(c.workspace, "src")
	if err = os.Symlink(src, target); err != nil {
		fmt.Printf("Failed to link directory, falling back to copy")

		if err = os.CopyFS(target, os.DirFS(src)); err != nil {
			return err
		}
	}

	fmt.Println("Writing descriptors")
	descContent := strings.Builder{}
	descContent.WriteString(fmt.Sprintf(`version="%s"
name="%s"
path="%s"
`, "1.0", name, target))
	if err = os.WriteFile(descA, []byte(descContent.String()), 0644); err != nil {
		return fmt.Errorf("failed to write descriptor: %v", err)
	}
	return nil
}
