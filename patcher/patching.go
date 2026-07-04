package patcher

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type Patcher struct {
	vanilla  string
	repoRoot string
	patches  string
	sources  string
	backups  string
	checker  FileChangeChecker
	dmp      diffmatchpatch.DiffMatchPatch
}

func NewPatcher(repo string, vanilla string) (*Patcher, error) {
	abs, err := filepath.Abs(repo)
	if err != nil {
		return nil, err
	}

	sources := filepath.Join(abs, "src")

	return &Patcher{
		vanilla:  vanilla,
		repoRoot: abs,
		sources:  sources,
		patches:  filepath.Join(abs, "patches"),
		backups:  filepath.Join(abs, "backups"),
		checker: *NewChangeChecker(
			sources,
			filepath.Join(abs, "metadata"),
		),
		dmp: *diffmatchpatch.New(),
	}, nil
}

func (p *Patcher) patchFile(file string) string {
	return filepath.Join(p.patches, file+".patch")
}

func (p *Patcher) sourceFile(file string) string {
	return filepath.Join(p.sources, file)
}

func (p *Patcher) makeBackup(file string) (string, error) {
	bk := filepath.Join(p.backups, file)

	if stat, err := os.Stat(bk); err == nil {
		if !stat.IsDir() {
			return bk, nil
		}
	}

	src := p.vanillaFile(file)
	return bk, copyFile(src, bk)
}

func (p *Patcher) vanillaFile(file string) string {
	return filepath.Join(p.vanilla, file)
}

func (p *Patcher) Apply(file string) error {
	vanilla, err := p.makeBackup(file)
	if err != nil {
		return fmt.Errorf("failed to create backup for %s: %v", file, err)
	}
	patch := p.patchFile(file)

	if stat, err := os.Stat(vanilla); err != nil {
		if os.IsNotExist(err) {
			return errors.New("vanilla file not found")
		}
		return err
	} else if stat.IsDir() {
		return errors.New("cannot apply patch for directory")
	}

	if stat, err := os.Stat(patch); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return copyFile(vanilla, p.sourceFile(file))
	} else if stat.IsDir() {
		return errors.New("cannot apply patch from directory")
	}

	// apply with git
	return p.applyPatch(vanilla, patch, p.sourceFile(file))
}

func (p *Patcher) ApplyAll() error {
	metadata := filepath.Join(p.repoRoot, "metadata")

	return filepath.WalkDir(p.patches, func(path string, d fs.DirEntry, wErr error) error {
		if wErr != nil {
			return wErr
		}
		if d != nil && d.IsDir() {
			if cDir, err := filepath.Abs(path); err == nil {
				if cDir == metadata {
					return filepath.SkipDir
				}
			} else {
				return err
			}
			return nil
		}

		if rel, err := filepath.Rel(p.patches, path); err != nil {
			return err
		} else {
			rel = rel[:len(rel)-6]
			fmt.Println("Applying " + rel)

			if err = p.Apply(rel); err != nil {
				return err
			}
		}

		return nil
	})
}

func (p *Patcher) SaveStates() error {
	return p.checker.SaveCurrentState(".")
}

func (p *Patcher) Generate(file string) error {
	src := p.sourceFile(file)
	if stat, err := os.Stat(src); err != nil {
		if os.IsNotExist(err) {
			return errors.New("source file not found")
		}
		return err
	} else if stat.IsDir() {
		return errors.New("cannot generate patch for directory")
	}

	vanilla, err := p.makeBackup(file)
	if err != nil {
		return fmt.Errorf("failed to create backup for %s: %v", file, err)
	}
	if stat, err := os.Stat(vanilla); err != nil {
		if os.IsNotExist(err) {
			return errors.New("vanilla file not found")
		}
		return err
	} else if stat.IsDir() {
		return errors.New("cannot generate patch for directory")
	}

	patch := p.patchFile(file)
	if err := p.generatePatch(vanilla, src, patch); err != nil {
		return err
	}

	return p.checker.SaveCurrentState(file)
}

func (p *Patcher) RegenerateChanged() error {
	changes, err := p.checker.ListChanges(".")
	if err != nil {
		return err
	}

	all := len(changes)
	for idx, file := range changes {
		fmt.Printf("Regenerating %s (%d/%d)\n", file, idx+1, all)
		if err = p.Generate(file); err != nil { // generate patch
			fmt.Println("Error regenerating patch for", file)
			return err
		}

		// save state
		if err = p.checker.SaveCurrentState(file); err != nil {
			fmt.Println("Error regenerating patch for", file)
			return err
		}
	}

	return nil
}
