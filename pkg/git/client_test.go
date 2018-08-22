package git_test

import (
	"errors"
	"os/exec"

	"github.com/loggregator/bumper/pkg/git"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	It("gets commits hashes for a given range", func() {
		se := &stubCommandExecutor{
			runResults: []runResult{
				{output: "f00dface\ndeadbeef\n123456\n789abc\ndef123\n"},
				{output: "Fifth Commit [Delivers #55555555]\n"},
				{output: "Fifth Commit [Delivers #55555555]\n"},
				{output: "Fourth Commit [fixes #44444444]\n"},
				{output: "Fourth Commit [fixes #44444444]\n"},
				{output: "Third Commit\n"},
				{output: "Third Commit\n\n[#33333333]\n"},
				{output: "Second Commit\n"},
				{output: "Second Commit\n\n[#22222222]\n"},
				{output: "First Commit\n"},
				{output: "First Commit\n\n[finishes #11111111]\n"},
			},
		}
		gc := git.NewClient(git.WithCommandExecutor(se))

		commits, err := gc.Commits("master..release-elect")
		Expect(err).ToNot(HaveOccurred())

		Expect(se.runCommands).To(HaveLen(11))
		Expect(se.runCommands[0].Args).To(Equal([]string{
			"git", "log", "--pretty=format:%H", "master..release-elect",
		}))

		Expect(se.runCommands[1].Args).To(Equal([]string{
			"git", "show", "--no-patch", "--pretty=format:%s", "f00dface",
		}))
		Expect(se.runCommands[2].Args).To(Equal([]string{
			"git", "show", "--pretty=format:%B", "f00dface",
		}))
		Expect(se.runCommands[3].Args).To(Equal([]string{
			"git", "show", "--no-patch", "--pretty=format:%s", "deadbeef",
		}))
		Expect(se.runCommands[4].Args).To(Equal([]string{
			"git", "show", "--pretty=format:%B", "deadbeef",
		}))
		Expect(se.runCommands[5].Args).To(Equal([]string{
			"git", "show", "--no-patch", "--pretty=format:%s", "123456",
		}))
		Expect(se.runCommands[6].Args).To(Equal([]string{
			"git", "show", "--pretty=format:%B", "123456",
		}))
		Expect(se.runCommands[7].Args).To(Equal([]string{
			"git", "show", "--no-patch", "--pretty=format:%s", "789abc",
		}))
		Expect(se.runCommands[8].Args).To(Equal([]string{
			"git", "show", "--pretty=format:%B", "789abc",
		}))
		Expect(se.runCommands[9].Args).To(Equal([]string{
			"git", "show", "--no-patch", "--pretty=format:%s", "def123",
		}))
		Expect(se.runCommands[10].Args).To(Equal([]string{
			"git", "show", "--pretty=format:%B", "def123",
		}))

		Expect(commits).To(HaveLen(5))
		Expect(commits).To(Equal([]*git.Commit{
			{Hash: "f00dface", Subject: "Fifth Commit [Delivers #55555555]\n", StoryID: 55555555},
			{Hash: "deadbeef", Subject: "Fourth Commit [fixes #44444444]\n", StoryID: 44444444},
			{Hash: "123456", Subject: "Third Commit\n", StoryID: 33333333},
			{Hash: "789abc", Subject: "Second Commit\n", StoryID: 22222222},
			{Hash: "def123", Subject: "First Commit\n", StoryID: 11111111},
		}))
	})

	It("gets commits of submodules", func() {
		se := &stubCommandExecutor{
			runResults: []runResult{
				{output: "123456\n789012\n"},
				{output: "Third Commit\n"},
				{output: "Bump src/bumper1\n\n  Username:\n    Update Bumper\n\n\n+Subproject commit ab321c"},
				{output: "Sub Commit\n\n[#44444444]"},
				{output: "Second Commit\n"},
				{output: "Bump src/bumper2\n\n  Username:\n    Update Bumper\n\n\n+Subproject commit cd432b"},
				{output: "Sub Commit\n\n[#55555555]"},
			},
		}
		gc := git.NewClient(
			git.WithCommandExecutor(se),
			git.WithFollowBumpsOf("src/bumper1", "src/bumper2"),
		)

		commits, err := gc.Commits("master..release-elect")
		Expect(err).ToNot(HaveOccurred())

		Expect(se.runCommands).To(HaveLen(7))
		Expect(se.runCommands[0].Args).To(Equal([]string{
			"git", "log", "--pretty=format:%H", "master..release-elect",
		}))

		Expect(se.runCommands[1].Args).To(Equal([]string{
			"git", "show", "--no-patch", "--pretty=format:%s", "123456",
		}))
		Expect(se.runCommands[2].Args).To(Equal([]string{
			"git", "show", "--pretty=format:%B", "123456",
		}))
		Expect(se.runCommands[3].Args).To(Equal([]string{
			"git", "-C", "src/bumper1", "show", "--no-patch", "--pretty=format:%B", "ab321c",
		}))
		Expect(se.runCommands[6].Args).To(Equal([]string{
			"git", "-C", "src/bumper2", "show", "--no-patch", "--pretty=format:%B", "cd432b",
		}))

		Expect(commits).To(HaveLen(2))
		Expect(commits).To(Equal([]*git.Commit{
			{Hash: "123456", Subject: "Third Commit\n", StoryID: 44444444},
			{Hash: "789012", Subject: "Second Commit\n", StoryID: 55555555},
		}))
	})

	It("returns an error if git log fails", func() {
		se := &stubCommandExecutor{
			runResults: []runResult{
				{err: errors.New("could not get log")},
			},
		}
		gc := git.NewClient(git.WithCommandExecutor(se))

		_, err := gc.Commits("master..release-elect")
		Expect(err).To(HaveOccurred())
	})

	It("returns an error if git show no-patch fails", func() {
		se := &stubCommandExecutor{
			runResults: []runResult{
				{output: "123456\n789abc\ndef123\n"},
				{err: errors.New("could not git show commit")},
			},
		}
		gc := git.NewClient(git.WithCommandExecutor(se))

		_, err := gc.Commits("master..release-elect")
		Expect(err).To(HaveOccurred())
	})

	It("returns an error if git show fails", func() {
		se := &stubCommandExecutor{
			runResults: []runResult{
				{output: "123456\n789abc\ndef123\n"},
				{output: "Third Commit\n"},
				{err: errors.New("could not git show commit")},
			},
		}
		gc := git.NewClient(git.WithCommandExecutor(se))

		_, err := gc.Commits("master..release-elect")
		Expect(err).To(HaveOccurred())
	})
})

type runResult struct {
	output string
	err    error
}

type stubCommandExecutor struct {
	runCommands []*exec.Cmd
	runResults  []runResult
}

func (s *stubCommandExecutor) Run(cmd *exec.Cmd) error {
	s.runCommands = append(s.runCommands, cmd)

	r := s.runResults[len(s.runCommands)-1]
	if r.err != nil {
		return r.err
	}

	cmd.Stdout.Write([]byte(r.output))

	return nil
}
