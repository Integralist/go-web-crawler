package crawler

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/integralist/go-web-crawler/internal/instrumentator"
	"github.com/integralist/go-web-crawler/internal/mapper"
	"github.com/integralist/go-web-crawler/internal/requester"
	"github.com/sirupsen/logrus"
)

// Tracker is a simplified version of sync.Map which will aid with testing.
type Tracker interface {
	Load(key interface{}) (value interface{}, ok bool)
	Store(key, value interface{})
}

const defaultWorkerPool = 20

// dot indicates whether we should be outputting any print information
var dot bool

// json indicates whether we should be outputting any print information.
var json bool

// log is a preconfigured logger instance.
var log *logrus.Entry

// httpClient is a preconfigured HTTP client.
var httpClient requester.HTTPClient

// red provides coloured output for text given to a string format function.
var red = color.New(color.FgRed).SprintFunc()

// green provides coloured output for text given to a string format function.
var green = color.New(color.FgGreen).SprintFunc()

// yellow provides coloured output for text given to a string format function.
var yellow = color.New(color.FgYellow).SprintFunc()

// Init configures the package from an outside mediator
func Init(j, d bool, hc requester.HTTPClient, instr *instrumentator.Instr) {
	log = instr.Logger
	json = j
	dot = d
	httpClient = hc
}

// Crawl concurrently requests URLs extracted from a slice of mapper.Page
func Crawl(mappedPage mapper.Page, trackedURLs Tracker) []requester.Page {
	toProcess := len(mappedPage.Anchors)

	// avoid printing to stdout if user has requested json/dot formatted output
	if !json && !dot {
		fmt.Println("-------------------------")
		fmt.Println(mappedPage.URL)
		fmt.Printf("Contains %s URLs to crawl\n", red(toProcess))
	}

	// if the page has no anchors associated within it, then we'll skip
	// processing the current page
	if toProcess < 1 {
		if !json && !dot {
			fmt.Printf("Crawled %s URLs %s\n\n", green("0"), green("(no pages requested)"))
		}
		return []requester.Page{}
	}

	var counter int
	var mutex = &sync.Mutex{}
	var wg sync.WaitGroup
	var pages []requester.Page

	// dynamically determine the worker pool size, we'll either set a default or
	// use a smaller value if the number of tasks is smaller than the default.
	workerPool := defaultWorkerPool
	if toProcess < defaultWorkerPool {
		workerPool = toProcess
	}

	startTime := time.Now()
	tasks := make(chan string, workerPool)

	// spin up our worker pool as goroutines awaiting tasks to be processed
	for i := 0; i < workerPool; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			for url := range tasks {
				page, err := requester.Get(url, httpClient)
				if err != nil {
					log.Warn(err)
					continue
				}
				trackedURLs.Store(url, true)
				counter++

				mutex.Lock()
				pages = append(pages, page)
				mutex.Unlock()
			}
		}(i)
	}

	for _, url := range mappedPage.Anchors {
		// check anchor within the given page hasn't already been processed
		if _, ok := trackedURLs.Load(url); !ok {
			tasks <- url
		}
	}

	// go routines stay 'open' and blocking this function from finishing until we
	// close the tasks channel
	close(tasks)

	wg.Wait()

	// we'll colourize the output so we can see at a glance what's happening...
	//
	// red: we request the full number of URLs
	// yellow: we requested less than expected (as duplicates were found)
	// green: we requested zero URLs (as duplicates were found)
	counterOut := red(strconv.Itoa(toProcess))
	msg := ""
	if counter < toProcess {
		counterOut = yellow(counter)
	}
	if counter == 0 {
		counterOut = green(counter)
		msg = green("(no pages requested)")
	}

	// avoid printing to stdout if user has requested json/dot formatted output
	if !json && !dot {
		fmt.Printf("Crawled %s URLs %s\n\n", counterOut, msg)
	}

	log.Debug("time spent crawling:", time.Since(startTime))

	return pages
}
