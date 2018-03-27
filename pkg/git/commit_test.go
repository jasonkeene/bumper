package git_test

import (
	"github.com/loggregator/bumper/pkg/git"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Commit", func() {
	Describe("ShortSHA", func() {
		It("returns the first 8 characters of the sha", func() {
			c := git.Commit{
				Hash: "abcdef0123456789",
			}

			Expect(c.ShortSHA()).To(Equal("abcdef01"))
		})

		It("returns an empty string for no hash", func() {
			c := git.Commit{}

			Expect(c.ShortSHA()).To(Equal(""))
		})
	})

	Describe("FormatSubject", func() {
		It("truncates the subject and adds ellipsis if > passed length", func() {
			c := git.Commit{
				Subject: "1234567890",
			}

			Expect(c.FormatSubject(7)).To(Equal("1234..."))
		})

		It("adds padding to the right if subject length is shorter than passed length", func() {
			c := git.Commit{
				Subject: "123",
			}

			Expect(c.FormatSubject(7)).To(Equal("123    "))
		})

		It("returns the subject when subject length == passed length", func() {
			c := git.Commit{
				Subject: "1234567890",
			}

			Expect(c.FormatSubject(10)).To(Equal("1234567890"))
		})
	})
})
