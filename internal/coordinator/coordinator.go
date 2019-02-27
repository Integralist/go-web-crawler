package coordinator

import (
	"fmt"
	"sync"
	"time"

	"github.com/integralist/go-web-crawler/internal/crawler"
	"github.com/integralist/go-web-crawler/internal/formatter"
	"github.com/integralist/go-web-crawler/internal/instrumentator"
	"github.com/integralist/go-web-crawler/internal/mapper"
	"github.com/integralist/go-web-crawler/internal/parser"
	"github.com/integralist/go-web-crawler/internal/requester"
)

// ProcessedResults are the final results slice containing all crawled pages.
type ProcessedResults []mapper.Page

// Start begins crawling the given website starting with the entry page.
func Start(protocol, hostname string, httpclient requester.HTTPClient, instr *instrumentator.Instr) ProcessedResults {
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
	trackedURLs := new(sync.Map)
	trackedURLs.Store(pageURL, true)

	// parse the requested page
	tokenizedPage := parser.Parse(page, instr)

	// map the tokenized page, and its assets
	mappedPage := mapper.Map(tokenizedPage)

	// results stores the final structure of crawled pages and their assets.
	var results []mapper.Page

	// now we have the initial page analysis, we'll process the anchors.
	//
	// note: as the `process` function is recursive, we need to pass a slice of
	// mapper.Page type.
	//
	// would be good to avoid the code smell of wrapping our single page instance
	// within a slice by maybe replacing the []T with variadic arguments, but
	// that is likely to result in other trade-offs.
	entryPage := ProcessedResults{mappedPage}
	results = process(entryPage, results, trackedURLs, instr)

	return results
}

// Results displays the final output for the program.
func Results(results []mapper.Page, json, dot bool, startTime time.Time) {
	if json {
		fmt.Println(formatter.Pretty(results))
	} else if dot {
		fmt.Println(formatter.Dot(results))
	} else {
		formatter.Standard(results, startTime)
	}
}

// process recursively calls itself and processes the next set of mapped pages.
func process(mappedPages ProcessedResults, results []mapper.Page, trackedURLs crawler.Tracker, instr *instrumentator.Instr) ProcessedResults {
	for _, page := range mappedPages {
		crawledPages := crawler.Crawl(page, trackedURLs, instr)
		tokenizedNestedPages := parser.ParseCollection(crawledPages, instr)
		mappedNestedPages := mapper.MapCollection(tokenizedNestedPages, instr)

		for _, mnp := range mappedNestedPages {
			results = append(results, mnp)
		}

		// reassign the results so we can return them up the stack back to the
		// original caller for final display
		results = process(mappedNestedPages, results, trackedURLs, instr)
	}

	return results
}
