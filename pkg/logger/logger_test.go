package logger_test

import (
	"bytes"

	"github.com/loggregator/bumper/pkg/git"
	"github.com/loggregator/bumper/pkg/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger", func() {
	var (
		buf *bytes.Buffer
		log *logger.Logger
	)

	BeforeEach(func() {
		buf = bytes.NewBuffer(nil)
		log = logger.NewLogger(
			logger.WithWriter(buf),
		)
	})

	It("does not log header or commit", func() {
		log.Header("master..release-elect")
		log.Commit(&git.Commit{
			Accepted: true,
			Hash:     "abc123fg",
		})

		Expect(buf.String()).To(BeEmpty())
	})

	It("logs footer", func() {
		log.Footer("the-footer")

		Expect(buf.String()).To(Equal("the-footer\n"))
	})
})
