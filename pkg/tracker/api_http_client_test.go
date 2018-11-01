package tracker_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/loggregator/bumper/pkg/tracker"
)

var _ = Describe("Tracker API HTTP Client", func() {
	It("sets the correct header", func() {
		spyHttpClient := newSpyHTTPRequestClient()
		apiHttpClient := tracker.NewAPIHTTPClient(spyHttpClient, "some-token")

		apiHttpClient.Get("some-url")

		Expect(spyHttpClient.trackerToken).To(Equal("some-token"))
	})
})

type spyHTTPRequestClient struct {
	trackerToken string
}

func newSpyHTTPRequestClient() *spyHTTPRequestClient {
	return &spyHTTPRequestClient{}
}

func (t *spyHTTPRequestClient) Do(r *http.Request) (*http.Response, error) {
	t.trackerToken = r.Header.Get("X-TrackerToken")

	return nil, nil
}
