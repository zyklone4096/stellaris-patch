package patcher

import (
	"bufio"
	"io"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type Ignore struct {
	matcher gitignore.Matcher
}

func LoadIgnoreRules(reader io.Reader) *Ignore {
	scanner := bufio.NewScanner(reader)

	var patterns []gitignore.Pattern
	for scanner.Scan() {
		patterns = append(patterns, gitignore.ParsePattern(scanner.Text(), nil))
	}

	return &Ignore{
		matcher: gitignore.NewMatcher(patterns),
	}
}

func (ign *Ignore) Match(dir bool, path ...string) bool {
	return ign.matcher.Match(path, dir)
}
