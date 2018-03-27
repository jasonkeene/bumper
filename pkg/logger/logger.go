package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/loggregator/bumper/pkg/git"
)

type Logger struct {
	writer io.Writer
}

func NewLogger(opts ...LoggerOption) *Logger {
	l := &Logger{
		writer: os.Stdout,
	}

	for _, o := range opts {
		o(l)
	}

	return l
}

func (l *Logger) Header(string) {}

func (l *Logger) Commit(*git.Commit) {}

func (l *Logger) Footer(bumpSHA string) {
	fmt.Fprintln(l.writer, bumpSHA)
}

type LoggerOption func(*Logger)

func WithWriter(w io.Writer) LoggerOption {
	return func(l *Logger) {
		l.writer = w
	}
}
