package patcher

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type FileChangeChecker struct {
	root     string
	metadata string
}

type fileMetadata struct {
	Size    int64
	ModTime int64
}

func NewChangeChecker(root string, metadata string) *FileChangeChecker {
	return &FileChangeChecker{
		root:     root,
		metadata: metadata,
	}
}

func (c *FileChangeChecker) metadataPath(file string) string {
	return filepath.Join(c.metadata, file)
}

func (c *FileChangeChecker) readMetadata(file string) (*fileMetadata, error) {
	md := c.metadataPath(file)
	if data, err := os.ReadFile(md); err == nil {
		var out fileMetadata
		if err = json.Unmarshal(data, &out); err == nil {
			return &out, nil
		}
		return nil, err
	} else {
		return nil, err
	}
}

func (c *FileChangeChecker) writeMetadata(file string, metadata fileMetadata) error {
	if data, err := json.Marshal(metadata); err == nil {
		md := c.metadataPath(file)
		if err = os.MkdirAll(filepath.Dir(md), 0755); err != nil {
			return err
		}

		if err = os.WriteFile(md, data, 0644); err == nil {
			return nil
		}
		return err
	} else {
		return err
	}
}

func (c *FileChangeChecker) SaveCurrentState(file string) error {
	stat, err := os.Stat(filepath.Join(c.root, file))
	if err != nil {
		return err
	}

	return c.writeMetadata(file, fileMetadata{
		Size:    stat.Size(),
		ModTime: stat.ModTime().UnixMilli(),
	})
}

func (c *FileChangeChecker) HasChanged(file string) (bool, error) {
	stat, err := os.Stat(filepath.Join(c.root, file))
	if err != nil {
		return true, err
	}

	md, err := c.readMetadata(file)
	if err != nil {
		return true, nil
	}

	return md.Size != stat.Size() || md.ModTime != stat.ModTime().UnixMilli(), nil
}

func (c *FileChangeChecker) ListChanges(dir string) ([]string, error) {
	baseDir := filepath.Join(c.root, dir)
	stat, err := os.Stat(baseDir)
	if err != nil {
		return make([]string, 0), err
	}

	if !stat.Mode().IsDir() {
		return make([]string, 0), errors.New("not a directory")
	}
	metadataDir, err := filepath.Abs(c.metadata)
	if err != nil {
		return make([]string, 0), err
	}

	paths := make(chan string)
	changed := make(chan string)
	errCh := make(chan error, 1)
	setErr := func(err error) {
		if err == nil {
			return
		}

		select {
		case errCh <- err:
		default:
		}
	}

	var workerWG sync.WaitGroup
	workers := 8

	workerWG.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer workerWG.Done()
			for p := range paths {
				rel, err := filepath.Rel(c.root, p)
				if err != nil {
					setErr(err)
					continue
				}

				change, err := c.HasChanged(rel)
				if change {
					changed <- rel
				}
				if err != nil {
					setErr(err)
				}
			}
		}()
	}

	var results []string
	var resultsWG sync.WaitGroup
	resultsWG.Add(1)
	go func() {
		defer resultsWG.Done()
		for item := range changed {
			results = append(results, item)
		}
	}()

	walkErr := filepath.WalkDir(baseDir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			setErr(walkErr)
			return walkErr
		}

		if d != nil && d.IsDir() {
			currentDir, err := filepath.Abs(p)
			if err != nil {
				setErr(err)
				return err
			}
			if currentDir == metadataDir {
				return filepath.SkipDir
			}
			return nil
		}

		paths <- p
		return nil
	})
	close(paths)
	workerWG.Wait()
	close(changed)
	resultsWG.Wait()
	sort.Strings(results)

	if walkErr != nil {
		return results, walkErr
	}

	select {
	case err := <-errCh:
		return results, err
	default:
		return results, nil
	}
}

func (c *FileChangeChecker) SaveCurrentStates(dir string) error {
	baseDir := filepath.Join(c.root, dir)
	stat, err := os.Stat(baseDir)
	if err != nil {
		return err
	}

	if !stat.Mode().IsDir() {
		return errors.New("not a directory")
	}

	metadataDir, err := filepath.Abs(c.metadata)
	if err != nil {
		return err
	}

	return filepath.WalkDir(baseDir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d != nil && d.IsDir() {
			currentDir, err := filepath.Abs(p)
			if err != nil {
				return err
			}
			if currentDir == metadataDir {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(c.root, p)
		if err != nil {
			return err
		}

		return c.SaveCurrentState(rel)
	})
}
