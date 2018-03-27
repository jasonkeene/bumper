package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/loggregator/bumper/pkg/git"
)

type VerboseLogger struct {
	writer io.Writer
}

func NewVerboseLogger(opts ...VerboseLoggerOption) *VerboseLogger {
	l := &VerboseLogger{
		writer: os.Stderr,
	}

	for _, o := range opts {
		o(l)
	}

	return l
}

func (l *VerboseLogger) Header(commitRange string) {
	fmt.Fprintln(
		l.writer,
		"Bumping the following range of commits:",
		extraRed(commitRange),
	)
	fmt.Fprintln(l.writer)
}

func (l *VerboseLogger) Commit(c *git.Commit) {
	storyID := fmt.Sprint(c.StoryID)
	if storyID == "0" {
		storyID = "~~~~~~~~~"
	}

	fmt.Fprintln(
		l.writer,
		formatAccepted(c),
		yellow(c.ShortSHA()),
		c.FormatSubject(40),
		blue(storyID),
		c.StoryName,
	)
}

func (l *VerboseLogger) Footer(bumpSHA string) {
	fmt.Fprintln(l.writer, "")

	if bumpSHA == "" {
		fmt.Fprintln(l.writer, "There are no commits to bump!")
		return
	}

	fmt.Fprintln(l.writer, "This is the commit you should bump to:", extraRed(bumpSHA))
}

type VerboseLoggerOption func(*VerboseLogger)

func WithWriter(w io.Writer) VerboseLoggerOption {
	return func(l *VerboseLogger) {
		l.writer = w
	}
}

func formatAccepted(c *git.Commit) string {
	if c.Accepted || c.StoryID == 0 {
		return green("✓")
	}

	return red("✗")
}

func red(s string) string {
	return "\033[202m" + s + "\033[0m"
}

func extraRed(s string) string {
	return "\033[222m" + s + "\033[0m"
}

func green(s string) string {
	return "\033[32m" + s + "\033[0m"
}

func blue(s string) string {
	return "\033[34m" + s + "\033[0m"
}

func yellow(s string) string {
	return "\033[33m" + s + "\033[0m"
}
