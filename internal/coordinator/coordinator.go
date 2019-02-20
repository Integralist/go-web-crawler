package coordinator

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/integralist/go-web-crawler/internal/crawler"
	"github.com/integralist/go-web-crawler/internal/formatter"
	"github.com/integralist/go-web-crawler/internal/mapper"
	"github.com/integralist/go-web-crawler/internal/parser"
	"github.com/integralist/go-web-crawler/internal/requester"
	"github.com/sirupsen/logrus"
)

// green provides coloured output for text given to a string format function.
var green = color.New(color.FgGreen).SprintFunc()

// trackedURLs enables us to avoid requesting pages already processed.
var trackedURLs sync.Map

// results stores the final structure of crawled pages and their assets.
var results []mapper.Page

// Init kick starts the configuration of various package level variables, then
// begins the process of concurrently crawling pages.
func Init(protocol, hostname string, subdomains []string, logger *logrus.Entry, json, dot bool) {
	startTime := time.Now()
	log := logger

	// the following http client configuration is passed around so that when we
	// make multiple GET requests we don't have to recreate the net/http client.
	httpClient := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	// ensure imported packages have the right configuration, which makes the
	// code easier to manage compared to injecting these objects as dependencies
	// into all function calls (as it makes the function signatures messy as well
	// as makes nested function calls tedious).
	//
	// the trade-off from an Init function to exported variables is that the
	// signature for Init is equally tedious when there are lots of things
	// requiring configuration from outside the package.
	crawler.Init(logger, json, dot, &httpClient)
	mapper.Init(logger)
	parser.Init(logger, protocol, hostname)
	parser.SetValidHosts(hostname, subdomains)
	requester.Init(logger)

	// request entrypoint web page
	pageURL := fmt.Sprintf("%s://%s", protocol, hostname)
	page, err := requester.Get(pageURL, &httpClient)
	if err != nil {
		log.Fatal(err)
	}

	if page.Status != 200 {
		log.Fatal("Non 200 for entry page")
	}

	// to prevent doubling up the processing of urls that have already been
	// handled, we'll use a a hash table for O(n) constant time lookups.
	trackedURLs.Store(pageURL, true)

	// parse the requested page
	tokenizedPage := parser.Parse(page)

	// map the tokenized page, and its assets
	mappedPage := mapper.Map(tokenizedPage)

	// now we have the initial page analysis, we'll process the anchors.
	//
	// note: as the `process` function is recursive, we need to pass a slice of
	// mapper.Page type.
	process([]mapper.Page{mappedPage})

	if json {
		fmt.Println(formatter.Pretty(results))
	} else if dot {
		fmt.Println(formatter.Dot(results))
	} else {
		fmt.Printf("-------------------------\n\nNumber of URLs crawled and processed: %s\n", green(len(results)))
		fmt.Printf("Time: %s\n", green(time.Since(startTime)))
	}
}

// process recursively calls itself and processes the next set of mapped pages.
func process(mappedPages []mapper.Page) {
	for _, page := range mappedPages {
		nestedPages := crawler.Crawl(page, &trackedURLs)
		tokenizedNestedPages := parser.ParseCollection(nestedPages)
		mappedNestedPages := mapper.MapCollection(tokenizedNestedPages)

		for _, mnp := range mappedNestedPages {
			results = append(results, mnp)
		}

		process(mappedNestedPages)
	}
}