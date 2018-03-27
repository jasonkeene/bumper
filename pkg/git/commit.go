package git

import "strings"

type Commit struct {
	Hash      string
	Subject   string
	StoryID   int
	StoryName string
	Accepted  bool
}

func (c *Commit) ShortSHA() string {
	if len(c.Hash) < 8 {
		return c.Hash
	}

	return c.Hash[0:8]
}

func (c *Commit) FormatSubject(length int) string {
	if len(c.Subject) <= length {
		return c.Subject + strings.Repeat(" ", length-len(c.Subject))
	}

	return c.Subject[0:length-3] + "..."
}
