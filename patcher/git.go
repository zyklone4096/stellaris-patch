package patcher

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	patchOriginalName = "original"
	patchModifiedName = "modified"
)

func gitGeneratePatch(original string, modified string, dest string) error {
	tmp, err := os.MkdirTemp("", "stellaris-patch-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	if err = copyFile(original, filepath.Join(tmp, patchOriginalName)); err != nil {
		return fmt.Errorf("copy original: %w", err)
	}
	if err = copyFile(modified, filepath.Join(tmp, patchModifiedName)); err != nil {
		return fmt.Errorf("copy modified: %w", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("git", "diff", "--no-index", patchOriginalName, patchModifiedName)
	cmd.Dir = tmp
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err = cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
			return fmt.Errorf("git diff: %w: %s", err, strings.TrimSpace(stderr.String()))
		}
	}

	if err = os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	return os.WriteFile(dest, stdout.Bytes(), 0644)
}

func gitApplyPatch(original string, patch string, dest string) error {
	data, err := os.ReadFile(patch)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return copyFile(original, dest)
	}

	tmp, err := os.MkdirTemp("", "stellaris-patch-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	if err = copyFile(original, filepath.Join(tmp, patchOriginalName)); err != nil {
		return fmt.Errorf("copy original: %w", err)
	}
	modified := filepath.Join(tmp, patchModifiedName)
	if err = copyFile(original, modified); err != nil {
		return fmt.Errorf("copy modified base: %w", err)
	}

	var stderr bytes.Buffer
	cmd := exec.Command("git", "apply", "--whitespace=nowarn", patch)
	cmd.Dir = tmp
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("git apply: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return copyFile(modified, dest)
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
