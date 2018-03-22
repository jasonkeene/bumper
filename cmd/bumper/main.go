package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/loggregator/bumper/pkg/bumper"
	"github.com/loggregator/bumper/pkg/git"
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

	b := bumper.New(*commitRange, *verbose,
		bumper.WithGitClient(gc),
		bumper.WithTrackerClient(tc),
	)
	sha, ok := b.FindBumpSHA()
	if ok {
		fmt.Println(sha)
	}
}

type cmdExecutor struct{}

func (c cmdExecutor) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}
