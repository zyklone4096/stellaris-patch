package patcher

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func (p *Patcher) IsInstalled() bool {
	_, err := os.Stat(filepath.Join(p.vanilla, ".sp_installed"))
	return !os.IsNotExist(err)
}

func (p *Patcher) HardInstall() error {
	if p.IsInstalled() {
		return errors.New("existing installation detected in current game directory, try uninstall first")
	}

	// check sources
	if _, err := os.Stat(p.sources); !os.IsNotExist(err) {
		return errors.New("no source file found, please apply patch first")
	}

	// create installed mark
	if f, err := os.Create(filepath.Join(p.vanilla, ".sp_installed")); err != nil {
		return fmt.Errorf("failed to create installed mark: %v", err)
	} else {
		_ = f.Close()
	}

	return filepath.WalkDir(p.sources, func(path string, d fs.DirEntry, wErr error) error {
		if !d.IsDir() {
			if wErr != nil {
				return wErr
			}

			if rel, err := filepath.Rel(p.sources, path); err != nil {
				// make backup
				if _, err = p.makeBackup(rel); err != nil {
					return fmt.Errorf("failed to make backup for %s: %v", rel, err)
				}

				// copy file
				return copyFile(path, p.vanillaFile(rel))
			}
		}
		return nil
	})
}

func (p *Patcher) HardPurge(force bool) error {
	if !force && !p.IsInstalled() {
		return errors.New("no installed mark found, try with forced mode or check game files integrity in Steam")
	}

	errored := false
	if err := filepath.WalkDir(p.sources, func(path string, d fs.DirEntry, wErr error) error {
		if !d.IsDir() {
			if rel, err := filepath.Rel(p.sources, path); err != nil {
				if _, err := os.Stat(p.extraFile(rel)); os.IsNotExist(err) {
					// file exists in base game, copy from backup
					if err = copyFile(filepath.Join(p.backups, rel), path); err != nil {
						fmt.Printf("Failed to recover file %s from backup: %v\n", rel, err)
						errored = true
					}
				} else {
					// file not exists in base game, delete
					if err = os.Remove(path); err != nil {
						fmt.Printf("Failed to delete file %s: %v", rel, err)
						errored = true
					}
				}
			} else {
				fmt.Printf("Failed to get relative path of %s: %v\n", path, err)
				errored = true
			}
		}
		return nil
	}); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	if err := os.Remove(filepath.Join(p.vanilla, ".sp_installed")); err != nil {
		fmt.Printf("Failed to remove installation mark")
		errored = true
	}

	if errored {
		fmt.Printf("Errors occurred in uninstallation progress, please validate game files integrity in Steam or do a clean reinstall")
	}

	return nil
}
