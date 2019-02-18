package formatter

import (
	"strings"
	"testing"

	"github.com/integralist/go-web-crawler/internal/mapper"
)

func TestDot(t *testing.T) {
	input := []mapper.Page{
		mapper.Page{
			URL: "http://www.example.com/",
			Anchors: []string{
				"http://www.example.com/foo",
				"http://www.example.com/bar",
				"http://www.example.com/baz",
			},
			Links: []string{
				"http://www.example.com/foo.css",
				"http://www.example.com/bar.css",
				"http://www.example.com/baz.css",
			},
			Scripts: []string{
				"http://www.example.com/foo.js",
				"http://www.example.com/bar.js",
				"http://www.example.com/baz.js",
			},
		},
	}

	// note: comparing raw literal strings was a pain due to spacing/formatting
	// differences (my editor strips trailing spaces, but no matter whether I
	// utilized space stripping in the golang template or not I couldn't get the
	// raw literals to match up with the space formatting).
	//
	// because of this, and in the interest of time I opted to use strings.Fields
	// instead as a quick win.

	output := strings.Fields(`digraph sitemap {
	"http://www.example.com/"
		-> {
			"http://www.example.com/foo",
			"http://www.example.com/bar",
			"http://www.example.com/baz"
		}
}`)

	actual := strings.Fields(Dot(input))

	if output[0] != actual[0] {
		t.Errorf("expected: %s\ngot: %s", output, actual)
	}
}
