package logger_test

import (
	"bytes"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/loggregator/bumper/pkg/git"
	"github.com/loggregator/bumper/pkg/logger"
)

var _ = Describe("VerboseLogger", func() {
	var (
		buf *bytes.Buffer
		vl  *logger.VerboseLogger
	)

	BeforeEach(func() {
		buf = bytes.NewBuffer(nil)
		vl = logger.NewVerboseLogger(
			logger.WithVerboseWriter(buf),
		)
	})

	Describe("Header", func() {
		It("logs the commit range header", func() {

			vl.Header("master..release-elect")
			Expect(strings.Split(buf.String(), "\n")).To(Equal([]string{
				"Bumping the following range of commits: \033[222mmaster..release-elect\033[0m",
				"",
				"",
			}))
		})
	})

	Describe("Commit", func() {
		It("logs the commit with ✓ when the story is accepted", func() {
			vl.Commit(&git.Commit{
				Hash:      "ABC123DEF456",
				Subject:   "Update bumper to be awesome",
				StoryID:   12345678,
				StoryName: "My awesome story name",
				Accepted:  true,
			})
			Expect(strings.Split(buf.String(), "\n")).To(Equal([]string{
				"\033[32m✓\033[0m \033[33mABC123DE\033[0m Update bumper to be awesome              \033[34m12345678\033[0m My awesome story name",
				"",
			}))
		})

		It("logs the commit with ✗ when commit is not accepted", func() {
			vl.Commit(&git.Commit{
				Hash:      "ABC123DEF456",
				Subject:   "Update bumper to be awesome",
				StoryID:   12345678,
				StoryName: "My awesome story name",
				Accepted:  false,
			})
			Expect(strings.Split(buf.String(), "\n")).To(Equal([]string{
				"\033[202m✗\033[0m \033[33mABC123DE\033[0m Update bumper to be awesome              \033[34m12345678\033[0m My awesome story name",
				"",
			}))
		})

		It("logs the commit with ✓ when there is no story ID", func() {
			vl.Commit(&git.Commit{
				Hash:     "ABC123DEF456",
				Subject:  "Update bumper to be awesome",
				StoryID:  0,
				Accepted: false,
			})
			Expect(strings.Split(buf.String(), "\n")).To(Equal([]string{
				"\033[32m✓\033[0m \033[33mABC123DE\033[0m Update bumper to be awesome              \033[34m~~~~~~~~~\033[0m ",
				"",
			}))
		})
	})

	Describe("Footer", func() {
		It("prints a message indicating the sha to bump to", func() {
			vl.Footer("abc123")

			Expect(strings.Split(buf.String(), "\n")).To(Equal([]string{
				"",
				"This is the commit you should bump to: \033[222mabc123\033[0m",
				"",
			}))
		})

		It("prints a message indicating no bump commits", func() {
			vl.Footer("")

			Expect(strings.Split(buf.String(), "\n")).To(Equal([]string{
				"",
				"There are no commits to bump!",
				"",
			}))
		})
	})
})
