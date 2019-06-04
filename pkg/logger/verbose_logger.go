package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/loggregator/bumper/pkg/git"
)

type VerboseLogger struct {
	writer       io.Writer
	disableColor bool
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
		l.extraRed(commitRange),
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
		l.formatAccepted(c),
		l.yellow(c.ShortSHA()),
		c.FormatSubject(40),
		l.blue(storyID),
		c.StoryName,
	)
}

func (l *VerboseLogger) Footer(bumpSHA string) {
	fmt.Fprintln(l.writer, "")

	if bumpSHA == "" {
		fmt.Fprintln(l.writer, "There are no commits to bump!")
		return
	}

	fmt.Fprintln(l.writer, "This is the commit you should bump to:", l.extraRed(bumpSHA))
}

type VerboseLoggerOption func(*VerboseLogger)

func WithVerboseWriter(w io.Writer) VerboseLoggerOption {
	return func(l *VerboseLogger) {
		l.writer = w
	}
}

func WithColorDisabled() VerboseLoggerOption {
	return func(l *VerboseLogger) {
		l.disableColor = true
	}
}

func (l *VerboseLogger) formatAccepted(c *git.Commit) string {
	if c.Accepted || c.StoryID == 0 {
		return l.green("✓")
	}

	return l.red("✗")
}

func (l *VerboseLogger) red(s string) string {
	if l.disableColor {
		return s
	}

	return "\033[202m" + s + "\033[0m"
}

func (l *VerboseLogger) extraRed(s string) string {
	if l.disableColor {
		return s
	}

	return "\033[222m" + s + "\033[0m"
}

func (l *VerboseLogger) green(s string) string {
	if l.disableColor {
		return s
	}

	return "\033[32m" + s + "\033[0m"
}

func (l *VerboseLogger) blue(s string) string {
	if l.disableColor {
		return s
	}

	return "\033[34m" + s + "\033[0m"
}

func (l *VerboseLogger) yellow(s string) string {
	if l.disableColor {
		return s
	}

	return "\033[33m" + s + "\033[0m"
}
