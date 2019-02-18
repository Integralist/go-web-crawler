package main

import (
	"flag"
	"os"
	"strings"

	"github.com/integralist/go-web-crawler/internal/coordinator"
	"github.com/sirupsen/logrus"
)

var (
	dot        *bool
	hostname   string
	httponly   *bool
	json       *bool
	log        *logrus.Entry
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

	// log configuration
	log = logrus.WithFields(logrus.Fields{
		"version":  version,
		"hostname": hostname,
	})
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
	log.Debug("STARTUP_SUCCESSFUL")

	protocol := "https"
	if *httponly {
		protocol = "http"
	}

	subdomainsParsed := strings.Split(subdomains, ",")

	// TODO: redesign initalization as large signatures are a code smell
	coordinator.Init(protocol, hostname, subdomainsParsed, log, *json, *dot)
}
