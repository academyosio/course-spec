package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	coursespec "github.com/academyosio/course-spec"
	"github.com/academyosio/course-spec/schema"
	"github.com/academyosio/course-spec/validate"
)

// errInvalid is returned when a course has errors, after they're printed.
var errInvalid = errors.New("the course has errors")

func newRootCommand(out io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:           "coursespec",
		Short:         "Check and list courses in the course-spec format",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(out)
	root.AddCommand(newValidateCommand(), newLessonsCommand(), newSchemaCommand())
	root.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		fmt.Fprintln(c.ErrOrStderr(), err)
		return err
	})
	return root
}

// courseDirs returns dir when it holds a course, else every course beneath.
func courseDirs(dir string) ([]string, error) {
	if _, err := os.Stat(filepath.Join(dir, coursespec.CourseFile)); err == nil {
		return []string{dir}, nil
	}
	dirs, err := coursespec.Find(dir)
	if err != nil {
		return nil, err
	}
	if len(dirs) == 0 {
		return nil, fmt.Errorf("no course.json in or beneath %s", dir)
	}
	return dirs, nil
}

func checkFormat(format string) error {
	if format != "text" && format != "json" {
		return fmt.Errorf("--format must be text or json, not %q", format)
	}
	return nil
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

type courseFindings struct {
	Dir      string             `json:"dir"`
	Errors   int                `json:"errors"`
	Warnings int                `json:"warnings"`
	Findings []validate.Finding `json:"findings"`
}

func newValidateCommand() *cobra.Command {
	var format string
	var opts validate.Options
	cmd := &cobra.Command{
		Use:   "validate <dir>",
		Short: "Check a course, or every course beneath a directory; exits non-zero on errors",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkFormat(format); err != nil {
				return report(cmd, err)
			}
			dirs, err := courseDirs(args[0])
			if err != nil {
				return report(cmd, err)
			}
			var results []courseFindings
			total := 0
			for _, d := range dirs {
				fs := validate.Course(d, opts)
				if fs == nil {
					fs = []validate.Finding{}
				}
				e, w := validate.Counts(fs)
				total += e
				results = append(results, courseFindings{Dir: d, Errors: e, Warnings: w, Findings: fs})
			}
			out := cmd.OutOrStdout()
			if format == "json" {
				if err := writeJSON(out, map[string]any{"courses": results}); err != nil {
					return err
				}
			} else {
				for _, r := range results {
					for _, f := range r.Findings {
						fmt.Fprintln(out, f.String())
					}
					fmt.Fprintf(out, "%s: %d errors, %d warnings\n", r.Dir, r.Errors, r.Warnings)
				}
			}
			if total > 0 {
				return errInvalid
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "text", "output format: text or json")
	cmd.Flags().BoolVar(&opts.StrictNarration, "strict-narration", false, "warn about characters TTS reads badly")
	return cmd
}

func newLessonsCommand() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "lessons <dir>",
		Short: "List sections, lessons, and quizzes with their publish status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkFormat(format); err != nil {
				return report(cmd, err)
			}
			dirs, err := courseDirs(args[0])
			if err != nil {
				return report(cmd, err)
			}
			var reports []coursespec.Report
			for _, d := range dirs {
				c, err := coursespec.Load(d)
				if err != nil {
					return report(cmd, fmt.Errorf("%w\nrun coursespec validate %s for every problem and its fix", err, d))
				}
				reports = append(reports, c.Status())
			}
			out := cmd.OutOrStdout()
			if format == "json" {
				return writeJSON(out, map[string]any{"courses": reports})
			}
			for _, r := range reports {
				printReport(out, r)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "text", "output format: text or json")
	return cmd
}

func printReport(w io.Writer, r coursespec.Report) {
	fmt.Fprintf(w, "%s %s (%s)\n", r.ID, r.Version, r.Title)
	for _, s := range r.Sections {
		state := ""
		if !s.Released {
			state = "  [planned]"
		}
		fmt.Fprintf(w, "  %s: %s%s\n", s.ID, s.Title, state)
		for _, l := range s.Lessons {
			extra := l.Type
			if l.Slides > 0 {
				extra = fmt.Sprintf("%d slides", l.Slides)
			}
			fmt.Fprintf(w, "    %-10s %s: %s (%s)\n", l.Status, l.ID, l.Title, extra)
		}
		if q := s.Quiz; q != nil {
			fmt.Fprintf(w, "    %-10s %s: %s (quiz, %d questions)\n", q.Status, q.ID, q.Title, q.Questions)
		}
	}
}

func newSchemaCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "schema [name]",
		Short: "Print an embedded JSON Schema: " + strings.Join(schema.Names(), ", "),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				for _, n := range schema.Names() {
					fmt.Fprintln(cmd.OutOrStdout(), n)
				}
				return nil
			}
			data, err := schema.Get(args[0])
			if err != nil {
				return report(cmd, err)
			}
			_, err = cmd.OutOrStdout().Write(data)
			return err
		},
	}
}

// report prints err for the person running the command and returns it.
func report(cmd *cobra.Command, err error) error {
	fmt.Fprintln(cmd.ErrOrStderr(), "error:", err)
	return err
}
