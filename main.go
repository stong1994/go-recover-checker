package main

import (
	"fmt"
	"go.uber.org/multierr"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	fset := token.NewFileSet()
	files, err := findFiles([]string{"./"})
	if err != nil {
		return
	}

}

func findFiles(patterns []string) (_ []string, err error) {
	files := make(map[string]struct{})
	for _, pat := range patterns {
		fs, findErr := findGoFiles(pat)
		if findErr != nil {
			err = multierr.Append(err, fmt.Errorf("enumerating Go files in %q: %v", pat, err))
			continue
		}

		for _, f := range fs {
			files[f] = struct{}{}
		}
	}

	sortedFiles := make([]string, 0, len(files))
	for f := range files {
		sortedFiles = append(sortedFiles, f)
	}
	sort.Strings(sortedFiles)

	return sortedFiles, err
}

func findGoFiles(path string) ([]string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	var files []string
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		mode := info.Mode()
		switch {
		case mode.IsRegular():
			if strings.HasSuffix(path, ".go") {
				files = append(files, path)
			}

		case mode.IsDir():
			base := filepath.Base(path)
			switch {
			case len(base) == 0,
				base[0] == '.',
				base[0] == '_',
				base == "vendor":
				return filepath.SkipDir
			}
		}

		return nil
	})

	return files, err
}
