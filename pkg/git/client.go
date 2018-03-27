package git

import (
	"bufio"
	"bytes"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var (
	storyID         = regexp.MustCompile(`\[#(\d+)\]`)
	submoduleCommit = regexp.MustCompile(`\+Subproject commit ([[:xdigit:]]+)\b`)
)

type CommandExecutor interface {
	Run(*exec.Cmd) error
}

type GitClient struct {
	exec           CommandExecutor
	submodulePaths []string
}

func NewClient(opts ...ClientOption) GitClient {
	c := GitClient{}

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

func (c GitClient) Commits(commitRange string) ([]*Commit, error) {
	buf := bytes.NewBuffer(nil)
	err := c.execute(buf, "git", "log", "--pretty=format:%H", commitRange)
	if err != nil {
		return nil, err
	}

	var commits []*Commit
	br := bufio.NewReader(buf)
	for {
		shaBytes, _, err := br.ReadLine()
		if err != nil {
			break
		}

		commit, err := c.buildCommit(string(shaBytes))
		if err != nil {
			return nil, err
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

func (c GitClient) execute(buf *bytes.Buffer, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = buf

	return c.exec.Run(cmd)
}

func (c GitClient) buildCommit(sha string) (*Commit, error) {
	subBuf := bytes.NewBuffer(nil)
	err := c.execute(subBuf, "git", "show", "--no-patch", "--pretty=format:%s", sha)
	if err != nil {
		return nil, err
	}

	idBuf := bytes.NewBuffer(nil)
	err = c.execute(idBuf, "git", "show", "--pretty=format:%B", sha)
	if err != nil {
		return nil, err
	}

	storyID := c.getStoryID(idBuf.String())
	commit := &Commit{
		Hash:    sha,
		Subject: subBuf.String(),
		StoryID: storyID,
	}

	for _, sp := range c.submodulePaths {
		if commit.StoryID == 0 {
			commit.StoryID = c.getBumpedStoryId(idBuf.String(), sp)
		}
	}

	return commit, nil
}

func (c GitClient) getStoryID(body string) int {
	result := storyID.FindStringSubmatch(body)
	if len(result) < 2 {
		return 0
	}
	storyID := result[1]
	id, err := strconv.Atoi(storyID)
	if err != nil {
		return 0
	}
	return id
}

func (c GitClient) getBumpedStoryId(commitMessage, followBumpOf string) int {
	if !strings.Contains(commitMessage, "Bump "+followBumpOf) {
		return 0
	}

	result := submoduleCommit.FindStringSubmatch(commitMessage)
	if len(result) < 2 {
		return 0
	}

	submoduleCommitHash := result[1]
	out := &bytes.Buffer{}
	c.execute(out, "git", "-C", followBumpOf, "show", "--no-patch", "--pretty=format:%B", submoduleCommitHash)
	submoduleCommitMessage := out.String()
	return c.getStoryID(submoduleCommitMessage)
}

type ClientOption func(c *GitClient)

func WithCommandExecutor(exec CommandExecutor) ClientOption {
	return func(c *GitClient) {
		c.exec = exec
	}
}

func WithFollowBumpsOf(submodulePaths ...string) ClientOption {
	return func(c *GitClient) {
		c.submodulePaths = submodulePaths
	}
}
