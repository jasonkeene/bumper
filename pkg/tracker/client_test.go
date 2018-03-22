package tracker_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/loggregator/bumper/pkg/tracker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	Describe("IsAccepted", func() {
		Context("when state is accepted", func() {
			It("returns true", func() {
				shc := &stubHTTPClient{
					getResponses: []httpResponse{
						{body: responseBody(1, "accepted"), code: 200},
					},
				}

				client := tracker.NewClient(
					tracker.WithHTTPClient(shc),
				)
				accepted := client.IsAccepted(1)
				Expect(accepted).To(BeTrue())
				Expect(shc.getURLs).To(HaveLen(1))
				Expect(shc.getURLs[0]).To(Equal("https://www.pivotaltracker.com/services/v5/stories/1"))
			})
		})

		Context("when story ID is 0", func() {
			It("returns true", func() {
				shc := &stubHTTPClient{}
				client := tracker.NewClient(
					tracker.WithHTTPClient(shc),
				)

				accepted := client.IsAccepted(0)
				Expect(accepted).To(BeTrue())
			})
		})

		Context("when state is not accepted", func() {
			It("returns false", func() {
				shc := &stubHTTPClient{
					getResponses: []httpResponse{
						{body: responseBody(1, "finished"), code: 200},
					},
				}

				client := tracker.NewClient(
					tracker.WithHTTPClient(shc),
				)
				accepted := client.IsAccepted(1)
				Expect(accepted).To(BeFalse())
				Expect(shc.getURLs).To(HaveLen(1))
				Expect(shc.getURLs[0]).To(Equal("https://www.pivotaltracker.com/services/v5/stories/1"))
			})
		})

	})

	Describe("Name", func() {
		It("returns the story name", func() {
			shc := &stubHTTPClient{
				getResponses: []httpResponse{
					{body: responseBody(1, "some state"), code: 200},
				},
			}
			client := tracker.NewClient(
				tracker.WithHTTPClient(shc),
			)

			Expect(client.Name(1)).To(Equal("Story Name"))
			Expect(shc.getURLs).To(HaveLen(1))
			Expect(shc.getURLs[0]).To(Equal("https://www.pivotaltracker.com/services/v5/stories/1"))
		})

		It("returns empty string if story ID is 0", func() {
			shc := &stubHTTPClient{}
			client := tracker.NewClient(
				tracker.WithHTTPClient(shc),
			)

			Expect(client.Name(0)).To(Equal(""))
		})
	})

	Describe("Story caching", func() {
		It("caches the story when IsAccepted called", func() {
			shc := &stubHTTPClient{
				getResponses: []httpResponse{
					{body: responseBody(1, "some state"), code: 200},
				},
			}
			client := tracker.NewClient(
				tracker.WithHTTPClient(shc),
			)

			client.IsAccepted(1)
			client.Name(1)

			Expect(shc.getURLs).To(HaveLen(1))
		})

		It("caches the story when Name called", func() {
			shc := &stubHTTPClient{
				getResponses: []httpResponse{
					{body: responseBody(1, "some state"), code: 200},
				},
			}
			client := tracker.NewClient(
				tracker.WithHTTPClient(shc),
			)

			client.Name(1)
			client.IsAccepted(1)

			Expect(shc.getURLs).To(HaveLen(1))
		})
	})
})

type httpResponse struct {
	body string
	code int
	err  error
}

type stubHTTPClient struct {
	getURLs      []string
	getResponses []httpResponse
}

func (s *stubHTTPClient) Get(url string) (*http.Response, error) {
	s.getURLs = append(s.getURLs, url)

	resp := s.getResponses[len(s.getURLs)-1]
	if resp.err != nil {
		return nil, resp.err
	}

	return &http.Response{
		StatusCode: resp.code,
		Body:       ioutil.NopCloser(strings.NewReader(resp.body)),
	}, nil
}

func responseBody(storyID int, state string) string {
	return fmt.Sprintf(responseBodyTemplate, storyID, state)
}

var (
	responseBodyTemplate = `{
		"id": %d,
		"current_state": "%s",
		"name": "Story Name"
	}`
)
