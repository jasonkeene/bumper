package main

import (
	"flag"
	"os"
	"os/exec"
	"strings"

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

	tc := tracker.NewClient()

	var log bumper.Logger
	if *verbose {
		log = logger.NewVerboseLogger()
	}

	b := bumper.New(*commitRange, log,
		bumper.WithGitClient(gc),
		bumper.WithTrackerClient(tc),
	)
	b.FindBumpSHA()
}

type cmdExecutor struct{}

func (c cmdExecutor) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}
