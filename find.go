package coursespec

import (
	"io/fs"
	"os"
	"path/filepath"
)

// Find returns every course directory at or beneath root: each directory
// holding a course.json. It skips .git, node_modules, and build
// directories, and doesn't look inside a course for more courses.
func Find(root string) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		switch d.Name() {
		case ".git", "node_modules", "build":
			if path != root {
				return filepath.SkipDir
			}
		}
		if exists(filepath.Join(path, CourseFile)) {
			dirs = append(dirs, path)
			return filepath.SkipDir
		}
		return nil
	})
	return dirs, err
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
