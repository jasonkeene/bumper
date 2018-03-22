package tracker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const urlTemplate = "https://www.pivotaltracker.com/services/v5/stories/%d"

type Client struct {
	cache      map[int]story
	httpClient HTTPClient
}

func NewClient(options ...Option) Client {
	c := Client{
		cache:      make(map[int]story),
		httpClient: http.DefaultClient,
	}
	for _, o := range options {
		o(&c)
	}
	return c
}

func (c Client) IsAccepted(storyID int) bool {
	if storyID == 0 {
		return true
	}

	s := c.story(storyID)

	return s.State == "accepted"
}

func (c Client) Name(storyID int) string {
	if storyID == 0 {
		return ""
	}

	s := c.story(storyID)

	return s.Name
}

func (c Client) story(storyID int) story {
	s, ok := c.cache[storyID]
	if ok {
		return s
	}

	resp, err := c.httpClient.Get(fmt.Sprintf(urlTemplate, storyID))
	if err != nil {
		log.Fatalf("failed to get story: %s", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&s)
	if err != nil {
		log.Fatalf("failed to unmarshal story: %s")
	}

	c.cache[storyID] = s

	return s
}

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

type Option func(*Client)

func WithHTTPClient(hc HTTPClient) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

type story struct {
	State string `json:"current_state"`
	Name  string `json:"name"`
}
