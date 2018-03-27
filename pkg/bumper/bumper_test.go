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
			nameResults:     []string{"", ""},
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

		b := bumper.New("master..release-elect", &spyLogger{},
			bumper.WithGitClient(sgc),
			bumper.WithTrackerClient(stc),
		)
		sha := b.FindBumpSHA()
		Expect(sha).To(Equal("789abc"))

		Expect(sgc.commitsRange).To(Equal("master..release-elect"))
		Expect(stc.acceptedRequests).To(ConsistOf(55555555, 88888888))
	})

	It("doesn't fail when there are no commits in the range", func() {
		stc := &spyTrackerClient{}
		sgc := &spyGitClient{}

		b := bumper.New("master..release-elect", &spyLogger{},
			bumper.WithGitClient(sgc),
			bumper.WithTrackerClient(stc),
		)

		sha := b.FindBumpSHA()
		Expect(sha).To(Equal(""))
	})

	It("logs information about commits", func() {
		sl := &spyLogger{}
		stc := &spyTrackerClient{
			acceptedResults: []bool{true, true},
			nameResults:     []string{"One", "Two"},
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

		b := bumper.New("master..release-elect", sl,
			bumper.WithGitClient(sgc),
			bumper.WithTrackerClient(stc),
		)

		_ = b.FindBumpSHA()

		Expect(sl.headerCommitRange).To(Equal("master..release-elect"))
		Expect(sl.commits).To(Equal([]*git.Commit{
			{
				Hash:      "123456",
				Subject:   "SecondCommit",
				StoryID:   55555555,
				StoryName: "One",
				Accepted:  true,
			},
			{
				Hash:      "789abc",
				Subject:   "FirstCommit",
				StoryID:   88888888,
				StoryName: "Two",
				Accepted:  true,
			},
		}))
		Expect(sl.bumpSHA).To(Equal("123456"))
	})

	It("doesn't bump commits that aren't fully accepted", func() {
		sl := &spyLogger{}
		stc := &spyTrackerClient{
			acceptedResults: []bool{true, true, false, true},
			nameResults:     []string{"", "", "", ""},
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

		b := bumper.New("master..release-elect", sl,
			bumper.WithGitClient(sgc),
			bumper.WithTrackerClient(stc),
		)
		sha := b.FindBumpSHA()
		Expect(sha).To(Equal(""))
		Expect(sl.footerCalled).To(BeTrue())
		Expect(sl.bumpSHA).To(Equal(""))

		Expect(sgc.commitsRange).To(Equal("master..release-elect"))
		Expect(stc.acceptedRequests).To(ConsistOf(88888888, 55555555, 88888888, 44444444))
	})

	It("does not log a commit sha if getting commits errors", func() {
		sgc := &spyGitClient{
			commitsError: errors.New("an error"),
		}
		sl := &spyLogger{}

		b := bumper.New("master..release-elect", sl,
			bumper.WithGitClient(sgc),
		)
		sha := b.FindBumpSHA()
		Expect(sha).To(Equal(""))

		Expect(sl.footerCalled).To(BeTrue())
		Expect(sl.bumpSHA).To(Equal(""))
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
	nameCallCount    int
	nameResults      []string
}

func (stc *spyTrackerClient) IsAccepted(storyID int) bool {
	stc.acceptedRequests = append(stc.acceptedRequests, storyID)
	return stc.acceptedResults[len(stc.acceptedRequests)-1]
}

func (stc *spyTrackerClient) Name(storyID int) string {
	stc.nameCallCount++

	return stc.nameResults[stc.nameCallCount-1]
}

type spyLogger struct {
	headerCommitRange string
	commits           []*git.Commit
	bumpSHA           string
	footerCalled      bool
}

func (s *spyLogger) Header(commitRange string) {
	s.headerCommitRange = commitRange
}

func (s *spyLogger) Commit(c *git.Commit) {
	s.commits = append(s.commits, c)
}

func (s *spyLogger) Footer(bumpSHA string) {
	s.footerCalled = true
	s.bumpSHA = bumpSHA
}
