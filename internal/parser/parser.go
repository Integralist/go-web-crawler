package parser

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/integralist/go-web-crawler/internal/instrumentator"
	"github.com/integralist/go-web-crawler/internal/requester"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

const defaultWorkerPool = 20

// log is a preconfigured logger instance and is set by the coordinator package.
var log *logrus.Entry

// protocol is the scheme the user has specified (HTTPS or HTTP)
var protocol string

// hostname is the top-level host the user has specified (minus the subdomain)
var hostname string

// ValidHosts is a map of valid URLs that are then used for inspecting the
// returned HTML from a HTTP GET request, and is set by the main package.
var ValidHosts map[string]bool

// Assets represents a collection of tokenized HTML elements.
type Assets []html.Token

// Page represents the tokenized elements of a HTML page.
type Page struct {
	Anchors Assets
	Links   Assets
	Scripts Assets
	URL     string
}

// Init configures the package from an outside mediator
func Init(p, h, s string, instr *instrumentator.Instr) {
	protocol = p
	hostname = h
	log = instr.Logger
	setValidHosts(h, s)
}

// setValidHosts sets map of valid URLs that are then used for inspecting the
// returned HTML from a HTTP GET request.
//
// The implementation optimizes for performance and so it uses a map data
// structure (i.e. hash table) instead of using regular expressions for pattern
// matching within the HTML tree.
//
// The trade-offs to this approach are a larger memory allocation for using the
// map and also slightly more complex code.
//
// Note: I considered returning interface instead of manual type, but opted for
// simpler code (wasn't sure there was any real benefit to an interface type)
func setValidHosts(hostname string, subdomains string) {
	validURLs := map[string]bool{}
	subdomainsParsed := strings.Split(subdomains, ",")

	for _, subdomain := range subdomainsParsed {
		dot := "."
		if subdomain == "" {
			dot = ""
		}
		url := fmt.Sprintf("%s%s%s", subdomain, dot, hostname)
		validURLs[url] = true
	}

	ValidHosts = validURLs
}

// Parse accepts a read http.Request body and tokenizes it. It will construct a
// page struct consisting of the anchors, links and scripts for the given page.
func Parse(page requester.Page) Page {
	var anchors []html.Token
	var links []html.Token
	var scripts []html.Token

	r := bytes.NewReader(page.Body)
	tz := html.NewTokenizer(r)

	for {
		tt := tz.Next()

		switch {
		case tt == html.ErrorToken:
			log.Debug("PARSER_EOF")

			return Page{
				URL:     page.URL,
				Anchors: anchors,
				Links:   links,
				Scripts: scripts,
			}
		case tt == html.StartTagToken:
			t := tz.Token()

			isAnchor := t.Data == "a"
			isLink := t.Data == "link"
			isScript := t.Data == "script"

			if isLink && canonical(t.Attr) {
				continue
			}

			if (isAnchor || isLink) && excludeInvalidURLs(&t, "href") {
				continue
			}

			if isScript && (missingScriptSrc(t.Attr) || excludeInvalidURLs(&t, "src")) {
				continue
			}

			if isAnchor {
				anchors = append(anchors, t)
			}

			if isLink {
				links = append(links, t)
			}

			if isScript {
				scripts = append(scripts, t)
			}
		}
	}
}

// ParseCollection concurrently parses a slice of requester.Page
func ParseCollection(pages []requester.Page) []Page {
	var mutex = &sync.Mutex{}
	var wg sync.WaitGroup
	var tokenizedPages []Page

	// dynamically determine the worker pool size, we'll either set a default or
	// use a smaller value if the number of tasks is smaller than the default.
	toProcess := len(pages)
	workerPool := defaultWorkerPool
	if toProcess < defaultWorkerPool {
		workerPool = toProcess
	}

	startTime := time.Now()
	tasks := make(chan requester.Page, workerPool)

	// spin up our worker pool as goroutines awaiting tasks to be processed
	for i := 0; i < workerPool; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			for page := range tasks {
				tokenizedPage := Parse(page)

				mutex.Lock()
				tokenizedPages = append(tokenizedPages, tokenizedPage)
				mutex.Unlock()
			}
		}(i)
	}

	for _, page := range pages {
		if page.Status != 200 {
			log.Debug("non 200 page:", page.URL)
			continue
		}
		tasks <- page
	}

	close(tasks)

	wg.Wait()
	log.Debug("time spent parsing:", time.Since(startTime))

	return tokenizedPages
}
