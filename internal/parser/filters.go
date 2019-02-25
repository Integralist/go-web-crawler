package parser

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

var imagePattern, _ = regexp.Compile("(?:doc|ico|pdf|gif|jpg|png)")

// we want to ignore urls with fragments and parsing external domains
func excludeInvalidURLs(token *html.Token, key string) bool {
	for i, a := range token.Attr {
		if a.Key == key {
			rawurl := a.Val
			log := log.WithFields(logrus.Fields{"url": rawurl})

			url, err := url.Parse(rawurl)
			if err != nil {
				log.Debug("URL_INVALID")
				return true
			}

			if imagePattern.MatchString(url.Path) {
				log.Debug("URL_INVALID")
				return true
			}

			// normalize the host information
			if url.Host == "" || url.Host == hostname {
				url.Host = fmt.Sprintf("www.%s", hostname)

				prefix := "/"
				if strings.HasPrefix(a.Val, "/") {
					prefix = ""
				}

				// note: using url.Path will remove any incidents where the same url
				// has a fragment causing the hash table key to change
				token.Attr[i].Val = fmt.Sprintf("%s://%s%s%s", protocol, url.Host, prefix, url.Path)
			}

			if _, ok := ValidHosts[url.Host]; !ok {
				log.Debug("URL_INVALID")
				return true
			}

			return false
		}
	}

	return true
}

// some script tags have inline code and aren't external references
func missingScriptSrc(attr []html.Attribute) bool {
	for _, a := range attr {
		if a.Key == "src" {
			return false
		}
	}
	return true
}

// a link can actually just contain rel='canonical' and not link to a css file
func canonical(attr []html.Attribute) bool {
	for _, a := range attr {
		if a.Key == "rel" && a.Val == "canonical" {
			return true
		}
	}
	return false
}
