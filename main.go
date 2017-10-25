package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var (
	commitRange string
	trackerKey  string
	verbose     bool
	reStoryID   = regexp.MustCompile(`\[#(\d+)\]`)
)

const urlTemplate = "https://www.pivotaltracker.com/services/v5/stories/%d"

type Commit struct {
	Hash    string
	Subject string
	// TODO: use a pointer to story here vs denormalizing
	StoryID   int
	StoryName string
	Accepted  bool
}

type Story struct {
	ID    int    `json:"id"`
	State string `json:"current_state"`
	Name  string `json:"name"`
}

var princeQuotes = []string{
	"A strong spirit transcends rules.",
	"You can always tell when the groove is working or not.",
	"Everyone has a rock bottom.",
	"The internet's completely over.",
	"So tonight we gonna party like it's 1999.",
}

func randomPrinceQuote() string {
	return princeQuotes[rand.Int()%len(princeQuotes)]
}

func main() {
	// validate tracker api key is set
	trackerKey = os.Getenv("TRACKER_KEY")
	if trackerKey == "" {
		log.Fatalf("Invalid Tracker Key")
	}

	// figure out ranges default to master..release-elect
	flag.Parse()
	printfVerbose("Bumping the following range of commits: %s\n\n", extraRed(commitRange))

	// collect all commits to check
	cmd := exec.Command("git", "log", "--pretty=format:%H", commitRange)
	out := &bytes.Buffer{}
	cmd.Stdout = out
	cmd.Stderr = &bytes.Buffer{}
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Unable to run command: %s\nstdout: %s\nstderr: %s", err, out, cmd.Stderr)
	}
	commits := getCommits(out)

	// bail out early if we have nothing to bump to
	if len(commits) == 0 {
		printlnVerbose("There are no commits to bump!")
		return
	}

	// get subjects
	getSubjects(commits)

	// parse story ids
	getStoryIDs(commits)

	// in parallel check those commits against the tracker api
	isAccepted(commits)

	// output commit informations
	var maxSubject int
	for _, c := range commits {
		if len(c.Subject) > maxSubject {
			maxSubject = len(c.Subject)
		}
	}
	for _, c := range commits {
		mark := red("✗")
		if c.Accepted {
			mark = green("✓")
		}
		if c.StoryID == 0 {
			mark = prince("✓")
		}
		subject := c.Subject
		if len(subject) > 50 {
			subject = subject[:47] + "..."
		}
		subject = padRight(subject, " ", maxSubject)
		storyID := strconv.Itoa(c.StoryID)
		if c.StoryID == 0 {
			storyID = prince(getDancer())
		}
		storyName := c.StoryName
		if c.StoryName == "" {
			storyName = prince(randomPrinceQuote())
		}
		printlnVerbose(mark, yellow(c.Hash[:8]), grey(subject), blue(storyID), grey(storyName))
	}
	printlnVerbose()

	// reverse the commits before we find the bump commit
	reversed := make([]*Commit, len(commits))
	for i, c := range commits {
		reversed[len(commits)-1-i] = c
	}
	commits = reversed

	// find bump commit and output it
	bumpHash := findBump(commits)
	if bumpHash == "" {
		printlnVerbose("There are no commits to bump!")
		return
	}
	printfVerbose("This is the commit you should bump to: ")
	if verbose {
		fmt.Println(extraRed(bumpHash))
		return
	}
	fmt.Println(bumpHash)
}

func getCommits(r io.Reader) []*Commit {
	commits := make([]*Commit, 0)
	br := bufio.NewReader(r)
	for {
		bytes, _, err := br.ReadLine()
		if err != nil {
			break
		}
		commit := &Commit{Hash: string(bytes)}
		commits = append(commits, commit)
	}
	return commits
}

func getSubjects(commits []*Commit) {
	for _, c := range commits {
		cmd := exec.Command("git", "show", "--no-patch", "--pretty=format:%s", c.Hash)
		out := &bytes.Buffer{}
		cmd.Stdout = out
		cmd.Stderr = &bytes.Buffer{}
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Unable to run command: %s\nstdout: %s\nstderr: %s", err, out, cmd.Stderr)
		}
		c.Subject = out.String()
	}
}

func getStoryIDs(commits []*Commit) {
	for _, c := range commits {
		cmd := exec.Command("git", "show", "--no-patch", "--pretty=format:%B", c.Hash)
		out := &bytes.Buffer{}
		cmd.Stdout = out
		cmd.Stderr = &bytes.Buffer{}
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Unable to run command: %s\nstdout: %s\nstderr: %s", err, out, cmd.Stderr)
		}
		c.StoryID = getStoryID(out.String())
	}
}

func getStoryID(body string) int {
	result := reStoryID.FindStringSubmatch(body)
	if len(result) < 2 {
		return 0
	}
	storyID := result[1]
	id, err := strconv.Atoi(storyID)
	if err != nil {
		return 0
	}
	return id
}

func getStory(id int, stories chan<- Story) {
	url := fmt.Sprintf(urlTemplate, id)
	printlnVerbose("getting url: ", url)

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var story Story
	err = json.Unmarshal(bytes, &story)
	if err != nil {
		printlnVerbose("invalid response from api: ", string(bytes))
		log.Fatal(err)
	}

	stories <- story
}

func isAccepted(commits []*Commit) {
	ids := make(map[int]struct{}, 0)
	for _, c := range commits {
		if c.StoryID == 0 {
			continue
		}
		ids[c.StoryID] = struct{}{}
	}

	stories := make(chan Story, len(ids))
	for id, _ := range ids {
		go getStory(id, stories)
	}

	for i := 0; i < len(ids); i++ {
		s := <-stories
		for _, c := range commits {
			if c.StoryID == 0 {
				c.Accepted = true
				continue
			}
			if c.StoryID == s.ID {
				c.StoryName = s.Name
				if s.State == "accepted" {
					c.Accepted = true
				}
			}
		}
	}
}

func findBump(commits []*Commit) string {
	invalid := make(map[int]bool)
	firstUnaccepted := -1
	bumpHash := ""

	// find invalid index
	for i, c := range commits {
		if !c.Accepted {
			firstUnaccepted = i
			break
		}
	}

	// return early if all stories are accepted
	if firstUnaccepted == -1 {
		// this shouldn't panic since len(commits) is always > 0
		return commits[len(commits)-1].Hash
	}

	// record invalid stories
	for _, c := range commits[firstUnaccepted:] {
		if c.Accepted && c.StoryID != 0 {
			invalid[c.StoryID] = true
		}
	}

	// find last commit that is accpeted and not invalid
	for _, c := range commits[:firstUnaccepted] {
		_, ok := invalid[c.StoryID]
		if ok {
			break
		}
		bumpHash = c.Hash
	}
	return bumpHash
}

func padRight(str, pad string, lenght int) string {
	for {
		str += pad
		if len(str) > lenght {
			return str[0:lenght]
		}
	}
}

func printlnVerbose(s ...interface{}) {
	if verbose {
		fmt.Println(s...)
	}
}

func printfVerbose(s string, x ...interface{}) {
	if verbose {
		fmt.Printf(s, x...)
	}
}

func red(s string) string {
	return "\033[38;5;202m" + s + "\033[0m"
}

func extraRed(s string) string {
	return "\033[38;5;222m" + s + "\033[0m"
}

func green(s string) string {
	return "\033[38;5;82m" + s + "\033[0m"
}

func blue(s string) string {
	return "\033[1;34m" + s + "\033[0m"
}

func yellow(s string) string {
	return "\033[33m" + s + "\033[0m"
}

func grey(s string) string {
	return "\033[38;5;242m" + s + "\033[0m"
}

func prince(s string) string {
	return "\033[38;5;92m" + s + "\033[0m"
}

var dancerToggle bool

func getDancer() string {
	if dancerToggle {
		dancerToggle = false
		return "┏ (･o･)┛♪"
	}
	dancerToggle = true
	return "♪┗ (･o･)┓"
}

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.StringVar(
		&commitRange,
		"commit-range",
		"master..release-elect",
		"Specifies the commit range to consider bumping.",
	)
	flag.BoolVar(
		&verbose,
		"verbose",
		false,
		"Output all the information.",
	)
}
