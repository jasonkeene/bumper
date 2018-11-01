package tracker

import "net/http"

type RequestClient interface {
	Do(*http.Request) (*http.Response, error)
}

type ApiHTTPClient struct {
	client   RequestClient
	apiToken string
}

func NewAPIHTTPClient(client RequestClient, apiToken string) *ApiHTTPClient {
	return &ApiHTTPClient{
		client:   client,
		apiToken: apiToken,
	}
}

func (c *ApiHTTPClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-TrackerToken", c.apiToken)

	return c.client.Do(req)
}
