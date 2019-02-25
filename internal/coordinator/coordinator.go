package coordinator

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/integralist/go-web-crawler/internal/crawler"
	"github.com/integralist/go-web-crawler/internal/formatter"
	"github.com/integralist/go-web-crawler/internal/instrumentator"
	"github.com/integralist/go-web-crawler/internal/mapper"
	"github.com/integralist/go-web-crawler/internal/parser"
	"github.com/integralist/go-web-crawler/internal/requester"
)

// trackedURLs enables us to avoid requesting pages already processed.
var trackedURLs sync.Map

// results stores the final structure of crawled pages and their assets.
var results []mapper.Page

// Init kick starts the configuration of various package level variables, then
// begins the process of concurrently crawling pages.
func Init(protocol, hostname, subdomains string, json, dot bool, httpclient requester.HTTPClient, instr *instrumentator.Instr) {
	subdomainsParsed := strings.Split(subdomains, ",")

	// ensure imported packages have the right configuration, which makes the
	// code easier to manage compared to injecting these objects as dependencies
	// into all function calls (as it makes the function signatures messy as well
	// as makes nested function calls tedious).
	//
	// the trade-off from an Init function to exported variables is that the
	// signature for Init is equally tedious when there are lots of things
	// requiring configuration from outside the package.
	crawler.Init(instr, json, dot, httpclient)
	mapper.Init(instr)
	parser.Init(instr, protocol, hostname)
	parser.SetValidHosts(hostname, subdomainsParsed)
	requester.Init(instr)
}

// Start begins crawling the given website starting with the entry page.
func Start(protocol, hostname string, httpclient requester.HTTPClient, instr *instrumentator.Instr) {
	// request entrypoint web page
	pageURL := fmt.Sprintf("%s://%s", protocol, hostname)
	page, err := requester.Get(pageURL, httpclient)
	if err != nil {
		instr.Logger.Fatal(err)
	}

	if page.Status != 200 {
		instr.Logger.Fatal("Non 200 for entry page")
	}

	// to prevent doubling up the processing of urls that have already been
	// handled, we'll use a a hash table for O(1) constant time lookups.
	trackedURLs.Store(pageURL, true)

	// parse the requested page
	tokenizedPage := parser.Parse(page)

	// map the tokenized page, and its assets
	mappedPage := mapper.Map(tokenizedPage)

	// now we have the initial page analysis, we'll process the anchors.
	//
	// note: as the `process` function is recursive, we need to pass a slice of
	// mapper.Page type.
	//
	// would be good to avoid the code smell of wrapping our single page instance
	// within a slice by maybe replacing the []T with variadic arguments, but
	// that is likely to result in other trade-offs.
	process([]mapper.Page{mappedPage})
}

// Results displays the final output for the program.
func Results(json, dot bool, startTime time.Time) {
	if json {
		fmt.Println(formatter.Pretty(results))
	} else if dot {
		fmt.Println(formatter.Dot(results))
	} else {
		formatter.Standard(results, startTime)
	}
}

// process recursively calls itself and processes the next set of mapped pages.
func process(mappedPages []mapper.Page) {
	for _, page := range mappedPages {
		crawledPages := crawler.Crawl(page, &trackedURLs)
		tokenizedNestedPages := parser.ParseCollection(crawledPages)
		mappedNestedPages := mapper.MapCollection(tokenizedNestedPages)

		for _, mnp := range mappedNestedPages {
			results = append(results, mnp)
		}

		process(mappedNestedPages)
	}
}
