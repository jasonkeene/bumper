package bumper_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/loggregator/bumper/pkg/bumper"
	"github.com/loggregator/bumper/pkg/git"
)

var _ = Describe("Bumper", func() {
	It("gets commits in a given range", func() {
		stc := &spyTrackerClient{
			acceptedResults: []bool{false, true},
		}
		sgc := &spyGitClient{
			commitsResult: []*git.Commit{
				{
					Hash:    "123456",
					Subject: "SecondCommit",
					StoryID: 55555555,
				},
				{
					Hash:    "789abc",
					Subject: "FirstCommit",
					StoryID: 88888888,
				},
			},
		}

		b := bumper.New("master..release-elect", false,
			bumper.WithGitClient(sgc),
			bumper.WithTrackerClient(stc),
		)
		sha, ok := b.FindBumpSHA()
		Expect(ok).To(BeTrue())
		Expect(sha).To(Equal("789abc"))

		Expect(sgc.commitsRange).To(Equal("master..release-elect"))
		Expect(stc.acceptedRequests).To(ConsistOf(55555555, 88888888))
	})

	It("doesn't bump commits that aren't fully accepted", func() {
		stc := &spyTrackerClient{
			acceptedResults: []bool{true, true, false, true},
		}
		sgc := &spyGitClient{
			commitsResult: []*git.Commit{
				{
					Hash:    "456789",
					Subject: "FourthCommit",
					StoryID: 44444444,
				},
				{
					Hash:    "def123",
					Subject: "ThirdCommit",
					StoryID: 88888888,
				},
				{
					Hash:    "123456",
					Subject: "SecondCommit",
					StoryID: 55555555,
				},
				{
					Hash:    "789abc",
					Subject: "FirstCommit",
					StoryID: 88888888,
				},
			},
		}

		b := bumper.New("master..release-elect", false,
			bumper.WithGitClient(sgc),
			bumper.WithTrackerClient(stc),
		)
		sha, ok := b.FindBumpSHA()
		Expect(ok).To(BeFalse())
		Expect(sha).To(Equal(""))

		Expect(sgc.commitsRange).To(Equal("master..release-elect"))
		Expect(stc.acceptedRequests).To(ConsistOf(88888888, 55555555, 88888888, 44444444))
	})

	It("returns false for okay if getting commits errors", func() {
		sgc := &spyGitClient{
			commitsError: errors.New("an error"),
		}

		b := bumper.New("master..release-elect", false,
			bumper.WithGitClient(sgc),
		)
		sha, ok := b.FindBumpSHA()
		Expect(ok).To(BeFalse())
		Expect(sha).To(Equal(""))
	})
})

type spyGitClient struct {
	commitsRange  string
	commitsResult []*git.Commit
	commitsError  error
}

func (s *spyGitClient) Commits(commitsRange string) ([]*git.Commit, error) {
	s.commitsRange = commitsRange
	return s.commitsResult, s.commitsError
}

type spyTrackerClient struct {
	acceptedRequests []int
	acceptedResults  []bool
}

func (stc *spyTrackerClient) IsAccepted(storyNumber int) bool {
	stc.acceptedRequests = append(stc.acceptedRequests, storyNumber)
	return stc.acceptedResults[len(stc.acceptedRequests)-1]
}
