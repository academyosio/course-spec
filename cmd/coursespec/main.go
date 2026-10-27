// Command coursespec checks and lists course directories in the course-spec
// format.
package main

import (
	"os"
)

func main() {
	if err := newRootCommand(os.Stdout).Execute(); err != nil {
		os.Exit(1)
	}
}
