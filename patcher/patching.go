package patcher

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type Patcher struct {
	vanilla  string
	repoRoot string
	patches  string
	sources  string
	backups  string
	extras   string
	checker  FileChangeChecker
	dmp      diffmatchpatch.DiffMatchPatch
	ign      *Ignore
}

func NewPatcher(repo string, vanilla string) (*Patcher, error) {
	abs, err := filepath.Abs(repo)
	if err != nil {
		return nil, err
	}

	sources := filepath.Join(abs, "src")

	ignore, err := os.OpenFile(filepath.Join(abs, ".spignore"), os.O_RDONLY, 0644)
	var ign *Ignore
	if err == nil {
		defer ignore.Close()
		ign = LoadIgnoreRules(ignore)
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	return &Patcher{
		vanilla:  vanilla,
		repoRoot: abs,
		sources:  sources,
		patches:  filepath.Join(abs, "patches"),
		backups:  filepath.Join(abs, "backups"),
		extras:   filepath.Join(abs, "extras"),
		checker: *NewChangeChecker(
			sources,
			filepath.Join(abs, "metadata"),
		),
		dmp: *diffmatchpatch.New(),
		ign: ign,
	}, nil
}

func (p *Patcher) patchFile(file string) string {
	return filepath.Join(p.patches, file+".patch")
}

func (p *Patcher) sourceFile(file string) string {
	return filepath.Join(p.sources, file)
}

func (p *Patcher) extraFile(file string) string {
	return filepath.Join(p.extras, file)
}

func (p *Patcher) makeBackup(file string) (string, error) {
	bk := filepath.Join(p.backups, file)

	if stat, err := os.Stat(bk); err == nil {
		if !stat.IsDir() {
			return bk, nil
		}
	}

	src := p.vanillaFile(file)
	if stat, err := os.Stat(src); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
	} else {
		if stat.IsDir() {
			return "", errors.New("base file is directory")
		}
	}

	return bk, copyFile(src, bk)
}

func (p *Patcher) vanillaFile(file string) string {
	return filepath.Join(p.vanilla, file)
}

func (p *Patcher) ignored(rel string, dir bool) bool {
	if p.ign == nil {
		return false
	}
	return p.ign.Match(dir, strings.Split(filepath.ToSlash(rel), "/")...)
}

func (p *Patcher) Apply(file string) error {
	vanilla, err := p.makeBackup(file)
	if err != nil {
		return fmt.Errorf("failed to create backup for %s: %v", file, err)
	}
	if vanilla == "" { // added file, copy from extras
		return copyFile(p.extraFile(file), p.sourceFile(file))
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

	if err := filepath.WalkDir(p.patches, func(path string, d fs.DirEntry, wErr error) error {
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
			if rel, err := filepath.Rel(p.patches, path); err == nil {
				if p.ignored(rel, true) {
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
			if p.ignored(rel, false) {
				fmt.Printf("Skipping ignored %s\n", rel)
				return nil
			}

			fmt.Printf("Applying %s\n", rel)

			if err = p.Apply(rel); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	// copy added files
	if err := filepath.WalkDir(p.extras, func(path string, d fs.DirEntry, wErr error) error {
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
			if rel, err := filepath.Rel(p.extras, path); err == nil {
				if p.ignored(rel, true) {
					return filepath.SkipDir
				}
			} else {
				return err
			}
			return nil
		}

		if rel, err := filepath.Rel(p.extras, path); err != nil {
			return err
		} else {
			if p.ignored(rel, false) {
				fmt.Printf("Skipping ignored %s\n", rel)
				return nil
			}

			if _, err := os.Stat(p.vanillaFile(rel)); os.IsNotExist(err) { // no conflicting file, continue
				fmt.Printf("Copying added %s\n", rel)
				return copyFile(filepath.Join(p.extras, rel), p.sourceFile(rel))
			}

			fmt.Printf("Conflicting file detected on added file %s, skipping\n", rel)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
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
	if vanilla == "" { // added file, copy to extras
		return copyFile(src, p.extraFile(file))
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

	filtered := changes[:0]
	for _, file := range changes {
		if p.ignored(file, false) {
			fmt.Printf("Skipping ignored %s\n", file)
			continue
		}
		filtered = append(filtered, file)
	}
	changes = filtered

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
