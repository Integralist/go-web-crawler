package mapper

// The thinking behind having a separate 'mapper' package, as opposed to doing
// part of this processing within the parser package, was that you might want
// to do other things with the parsed/tokenized page.
//
// The mapper package can then focus specifically on filtering out elements
// that it is interested in, leaving the originally tokenized page data to be
// passed onto other packages that can either use the full dataset or filter
// out different fields.

import (
	"sync"
	"time"

	"github.com/integralist/go-web-crawler/internal/instrumentator"
	"github.com/integralist/go-web-crawler/internal/parser"
	"github.com/sirupsen/logrus"
)

const defaultWorkerPool = 20

// Assets represents a collection of related filtered HTML elements.
type Assets []string

// Page represents the filtered elements of a HTML page (anchors/links/scripts).
type Page struct {
	Anchors Assets
	Links   Assets
	Scripts Assets
	URL     string
}

// log is a preconfigured logger instance and is set by the coordinator package.
var log *logrus.Entry

// the mapper is executed concurrently, so we need appends to be thread-safe.
var mutex = &sync.Mutex{}

// Init configures the package from an outside mediator
func Init(instr *instrumentator.Instr) {
	log = instr.Logger
}

// Map associates static assets with its parent web page.
//
// This function is expected to be executed concurrently, and so we wrap the
// slice append calls with a mutex.
func Map(page parser.Page) Page {
	var trackedURLs sync.Map
	var anchors Assets
	var links Assets
	var scripts Assets

	anchors = appendWhenNotTracked("href", anchors, page.Anchors, &trackedURLs)
	links = appendWhenNotTracked("href", links, page.Links, &trackedURLs)
	scripts = appendWhenNotTracked("src", scripts, page.Scripts, &trackedURLs)

	return Page{
		URL:     page.URL,
		Anchors: anchors,
		Links:   links,
		Scripts: scripts,
	}
}

// a single page can repeatedly link to the same URL, so we don't bother
// appending those URLs more than once (this makes reading the final JSON
// output much cleaner).
func appendWhenNotTracked(key string, collection Assets, assets parser.Assets, trackedURLs *sync.Map) Assets {
	for _, pageAssets := range assets {
		for _, attr := range pageAssets.Attr {
			if attr.Key == key {
				// a single page can repeatedly link to the same URL, so we don't
				// bother appending those URLs more than once.
				if _, ok := trackedURLs.Load(attr.Val); !ok {
					mutex.Lock()
					collection = append(collection, attr.Val)
					mutex.Unlock()
					trackedURLs.Store(attr.Val, true)
				}
			}
		}
	}
	return collection
}

// MapCollection concurrently maps a slice of parser.Page
func MapCollection(pages []parser.Page) []Page {
	var wg sync.WaitGroup
	var mappedPages []Page

	// dynamically determine the worker pool size, we'll either set a default or
	// use a smaller value if the number of tasks is smaller than the default.
	toProcess := len(pages)
	workerPool := defaultWorkerPool
	if toProcess < defaultWorkerPool {
		workerPool = toProcess
	}

	startTime := time.Now()
	tasks := make(chan parser.Page, workerPool)

	// spin up our worker pool as goroutines awaiting tasks to be processed
	for i := 0; i < workerPool; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			for page := range tasks {
				mappedPage := Map(page)

				mutex.Lock()
				mappedPages = append(mappedPages, mappedPage)
				mutex.Unlock()
			}
		}(i)
	}

	for _, page := range pages {
		tasks <- page
	}

	close(tasks)

	wg.Wait()
	log.Debug("time spent mapping:", time.Since(startTime))

	return mappedPages
}
