package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"log"

	"github.com/loggregator/bumper/pkg/bumper"
	"github.com/loggregator/bumper/pkg/git"
	"github.com/loggregator/bumper/pkg/logger"
	"github.com/loggregator/bumper/pkg/tracker"
)

func main() {
	commitRange := flag.String(
		"commit-range",
		"master..release-elect",
		"Specifies the commit range to consider bumping.",
	)
	verbose := flag.Bool(
		"verbose",
		false,
		"Output all the information.",
	)

	flag.Parse()

	var submodulePaths []string
	followBumpsOf := os.Getenv("FOLLOW_BUMPS_OF")
	if followBumpsOf != "" {
		submodulePaths = strings.Split(followBumpsOf, ",")
	}

	gc := git.NewClient(
		git.WithCommandExecutor(cmdExecutor{}),
		git.WithFollowBumpsOf(submodulePaths...),
	)

	var httpClient tracker.HTTPClient = http.DefaultClient

	apiToken := os.Getenv("TRACKER_API_TOKEN")
	if apiToken != "" {
		httpClient = tracker.NewAPIHTTPClient(http.DefaultClient, apiToken)
	}

	tc := tracker.NewClient(tracker.WithHTTPClient(httpClient))

	var bumperLog bumper.Logger = logger.NewLogger()
	if *verbose {
		bumperLog = logger.NewVerboseLogger()
	}

	b := bumper.New(*commitRange, bumperLog,
		bumper.WithGitClient(gc),
		bumper.WithTrackerClient(tc),
	)
	err := b.FindBumpSHA()
	if err != nil {
		log.Fatal(err)
	}
}

type cmdExecutor struct{}

func (c cmdExecutor) Run(cmd *exec.Cmd) error {
	stderrBuf := &strings.Builder{}
	cmd.Stderr = stderrBuf
	err := cmd.Run()

	if err != nil {
		return fmt.Errorf(
			`failed to execute "%s": %s (stderr: "%s")`,
			strings.Join(cmd.Args, " "),
			err,
			strings.TrimRight(stderrBuf.String(), "\r\n"),
		)
	}

	return nil
}
