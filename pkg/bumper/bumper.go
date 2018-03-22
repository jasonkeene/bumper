package bumper

import (
	"github.com/loggregator/bumper/pkg/git"
)

type GitClient interface {
	Commits(commitRange string) ([]*git.Commit, error)
}

type TrackerClient interface {
	IsAccepted(storyNumber int) bool
}

type Bumper struct {
	commitRange string
	verbose     bool
	gc          GitClient
	tc          TrackerClient
}

func New(commitRange string, verbose bool, opts ...BumperOption) Bumper {
	b := Bumper{
		commitRange: commitRange,
		verbose:     verbose,
	}

	for _, o := range opts {
		o(&b)
	}

	return b
}

func (b Bumper) FindBumpSHA() (string, bool) {
	commitsDesc, err := b.gc.Commits(b.commitRange)
	if err != nil {
		return "", false
	}

	for _, c := range commitsDesc {
		c.Accepted = b.tc.IsAccepted(c.StoryID)
	}

	return findBump(reverse(commitsDesc))
}

func reverse(commits []*git.Commit) []*git.Commit {
	reversed := make([]*git.Commit, len(commits))
	for i, c := range commits {
		reversed[len(commits)-1-i] = c
	}
	return reversed
}

func findBump(commits []*git.Commit) (string, bool) {
	invalid := make(map[int]bool)
	firstUnaccepted := -1
	bumpHash := ""

	// find invalid index
	for i, c := range commits {
		if !c.Accepted {
			firstUnaccepted = i
			break
		}
	}

	// return early if all stories are accepted
	if firstUnaccepted == -1 {
		// this shouldn't panic since len(commits) is always > 0
		return commits[len(commits)-1].Hash, true
	}

	// record invalid stories
	for _, c := range commits[firstUnaccepted:] {
		if c.Accepted && c.StoryID != 0 {
			invalid[c.StoryID] = true
		}
	}

	// find last commit that is accpeted and not invalid
	for _, c := range commits[:firstUnaccepted] {
		_, ok := invalid[c.StoryID]
		if ok {
			break
		}
		bumpHash = c.Hash
	}

	return bumpHash, bumpHash != ""
}

type BumperOption func(b *Bumper)

func WithGitClient(gc GitClient) BumperOption {
	return func(b *Bumper) {
		b.gc = gc
	}
}

func WithTrackerClient(tc TrackerClient) BumperOption {
	return func(b *Bumper) {
		b.tc = tc
	}
}
