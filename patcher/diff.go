package patcher

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func (p *Patcher) generatePatch(original string, modified string, dest string) error {
	cmd := exec.Command("diff", "-u", original, modified)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err == nil {
		_ = os.MkdirAll(filepath.Dir(dest), 0755)
		var file *os.File
		if file, err = os.OpenFile(dest, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644); err == nil {
			defer file.Close()
			if _, err = io.Copy(file, stdout); err == nil {
				err = cmd.Wait()
				if err != nil {
					if exit, ok := errors.AsType[*exec.ExitError](err); ok {
						code := exit.ExitCode()
						if code != 0 && code != 1 {
							return err
						}
						return nil
					}
					return err
				}
			}
		}

		return err
	} else {
		if exit, ok := errors.AsType[*exec.ExitError](err); ok {
			code := exit.ExitCode()
			if code != 0 && code != 1 {
				return err
			}
			return nil
		}
		return err
	}
}

func (p *Patcher) applyPatch(original string, patch string, dest string) error {
	cmd := exec.Command("patch", original, patch, "-o", dest)
	if err := cmd.Run(); err == nil {
		return nil
	} else {
		if exit, ok := errors.AsType[*exec.ExitError](err); ok {
			code := exit.ExitCode()
			if code != 0 && code != 1 {
				return err
			}
			return nil
		}
		return err
	}
}

func copyFile(src string, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
