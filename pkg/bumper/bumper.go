package bumper

import (
	"github.com/loggregator/bumper/pkg/git"
)

type GitClient interface {
	Commits(commitRange string) ([]*git.Commit, error)
}

type TrackerClient interface {
	IsAccepted(storyID int) bool
	Name(storyID int) string
}

type Logger interface {
	Header(commitRange string)
	Commit(c *git.Commit)
	Footer(bumpSHA string)
}

type Bumper struct {
	commitRange string
	gc          GitClient
	tc          TrackerClient
	log         Logger
}

func New(commitRange string, log Logger, opts ...BumperOption) Bumper {
	b := Bumper{
		commitRange: commitRange,
		log:         log,
	}

	for _, o := range opts {
		o(&b)
	}

	return b
}

func (b Bumper) FindBumpSHA() string {
	b.log.Header(b.commitRange)

	commitsDesc, err := b.gc.Commits(b.commitRange)
	if err != nil {
		b.log.Footer("")
		return ""
	}

	for _, c := range commitsDesc {
		c.Accepted = b.tc.IsAccepted(c.StoryID)
		c.StoryName = b.tc.Name(c.StoryID)
	}

	for _, c := range commitsDesc {
		b.log.Commit(c)
	}

	sha := findBump(reverse(commitsDesc))
	b.log.Footer(sha)
	return sha
}

func reverse(commits []*git.Commit) []*git.Commit {
	reversed := make([]*git.Commit, len(commits))
	for i, c := range commits {
		reversed[len(commits)-1-i] = c
	}
	return reversed
}

func findBump(commits []*git.Commit) string {
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
		return commits[len(commits)-1].Hash
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

	return bumpHash
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
