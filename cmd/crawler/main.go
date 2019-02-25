package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/integralist/go-web-crawler/internal/coordinator"
	"github.com/integralist/go-web-crawler/internal/crawler"
	"github.com/integralist/go-web-crawler/internal/instrumentator"
	"github.com/integralist/go-web-crawler/internal/mapper"
	"github.com/integralist/go-web-crawler/internal/parser"
	"github.com/integralist/go-web-crawler/internal/requester"
	"github.com/sirupsen/logrus"
)

// instr contains pre-configured instrumentation tools
var instr instrumentator.Instr

var (
	dot        *bool
	hostname   string
	httponly   *bool
	json       *bool
	subdomains string
	version    string // set via -ldflags in Makefile
)

func init() {
	// instrumentation
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetReportCaller(true) // TODO: benchmark for performance implications

	// flag configuration
	dot = flag.Bool("dot", false, "returns dot format file for use with graphviz")
	httponly = flag.Bool("httponly", false, "indicates HTTPS vs HTTP")
	json = flag.Bool("json", false, "returns raw site structure JSON for the output")
	const (
		flagHostnameValue   = "integralist.co.uk"
		flagHostnameUsage   = "hostname to crawl"
		flagSubdomainsValue = "www,"
		flagSubdomainsUsage = "valid subdomains"
	)
	flag.StringVar(&hostname, "hostname", flagHostnameValue, flagHostnameUsage)
	flag.StringVar(&hostname, "h", flagHostnameValue, flagHostnameUsage+" (shorthand)")
	flag.StringVar(&subdomains, "subdomains", flagSubdomainsValue, flagSubdomainsUsage)
	flag.StringVar(&subdomains, "s", flagSubdomainsValue, flagSubdomainsUsage+" (shorthand)")
	flag.Parse()

	// instrumentation configuration
	//
	// we would in a real-world application configure this with additional fields
	// such as `Metric` (for handling the recording of metrics using a service
	// such as Datadog, just as an example)
	//
	// note: I prefer to configure instrumentation within the init function of
	// the main package, but because I'm then passing this struct instance around
	// to other functions in other packages, it means I need to use an exported
	// reference from a mediator package (i.e. the instrumentator package)
	instr = instrumentator.Instr{
		Logger: logrus.WithFields(logrus.Fields{
			"version":  version,
			"hostname": hostname,
		}),
	}
}

func main() {
	// note: I like log messages to be a bit more structured so I typically opt
	// for a format such as 'VERB_STATE' and 'NOUN_STATE' (as this makes searching
	// for errors within a log aggregator easier).
	//
	// note: I typically prefer the "no news is good news" approach: which is
	// where you only log errors or warnings (not info/debug), as that makes
	// debugging easier as you don't have to filter out pointless messages about
	// things you already expected to happen, and the logs can instead focus on
	// surfacing all the _unexpected_ things that happened.
	instr.Logger.Debug("STARTUP_SUCCESSFUL")

	protocol := "https"
	if *httponly {
		protocol = "http"
	}

	// the following http client configuration is passed around so that when we
	// make multiple GET requests we don't have to recreate the net/http client.
	httpClient := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	// we will time how long our program takes to run.
	startTime := time.Now()

	// TODO: remove these Init calls altogether
	// initialize our packages with the relevant configuration
	crawler.Init(*json, *dot, &httpClient, &instr)
	mapper.Init(&instr)
	parser.Init(protocol, hostname, subdomains, &instr)
	requester.Init(&instr)

	// TODO: redesign initalization as large signatures are a code smell
	results := coordinator.Start(protocol, hostname, &httpClient, &instr)
	coordinator.Results(results, *json, *dot, startTime)
}
