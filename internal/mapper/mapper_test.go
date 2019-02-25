package mapper

import (
	"testing"

	"github.com/integralist/go-web-crawler/internal/instrumentator"
	"github.com/integralist/go-web-crawler/internal/parser"
	"github.com/integralist/go-web-crawler/internal/requester"
	"github.com/sirupsen/logrus"
)

/*
although TestMap is intended to test just the mapper.Map function, I decided to
call the parser.Parse function as part of the test setup. The rationale here is
that constructing a requester.Page instance is easier than trying to construct
multiple html.Token struct instances. the trade-off here is that we need to
ensure the parser package is configured properly up front, so there's slightly
more overhead in running the tests, and it starts to mimic more an
'integration' test than a simple 'unit' test.
*/

func TestMap(t *testing.T) {
	// we need to ensure a logger is initialized
	logger := instrumentator.Instr{
		Logger: logrus.NewEntry(logrus.New()),
	}
	parser.Init(&logger, "http", "example.com")

	// also need to ensure we tell the parser what hosts are valid
	parser.SetValidHosts("example.com", "www")

	// notice 'foo' assets appear twice but should be filtered out by the mapper
	// so that there is only one of them for each type (link/script).
	// also the canonical link should be filtered out too.
	page := requester.Page{
		URL: "http://www.example.com",
		Body: []byte(`<html>
	<head>
		<link rel='canonical' href='http://www.example.com/'>
		<link href="foo.css">
		<link href="foo.css">
		<link href="bar.css">
		<link href="baz.css">
		<script src="foo.js"></script>
		<script src="foo.js"></script>
		<script src="bar.js"></script>
		<script src="baz.js"></script>
	</head>
	<body>
		<a href="/foo">foo 1</a>
		<a href="/foo">foo 2</a>
		<a href="/bar">bar</a>
		<a href="/baz">baz</a>
	</body>
</html>`),
		Status: 200,
	}

	input := parser.Parse(page)

	output := Page{
		URL: "http://www.example.com",
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
	}

	actual := Map(input)

	if actual.URL != output.URL {
		t.Errorf("expected: %+v\ngot: %+v", output.URL, actual.URL)
	}

	if len(actual.Anchors) != len(output.Anchors) {
		t.Errorf("expected: %+v\ngot: %+v", len(output.Anchors), len(actual.Anchors))
	}

	if len(actual.Links) != len(output.Links) {
		t.Errorf("expected: %+v\ngot: %+v", len(output.Links), len(actual.Links))
	}

	if len(actual.Scripts) != len(output.Scripts) {
		t.Errorf("expected: %+v\ngot: %+v", len(output.Scripts), len(actual.Scripts))
	}

	for i, v := range actual.Anchors {
		if v != output.Anchors[i] {
			t.Errorf("expected: %+v\ngot: %+v", output.Anchors[i], v)
		}
	}

	for i, v := range actual.Links {
		if v != output.Links[i] {
			t.Errorf("expected: %+v\ngot: %+v", output.Links[i], v)
		}
	}

	for i, v := range actual.Scripts {
		if v != output.Scripts[i] {
			t.Errorf("expected: %+v\ngot: %+v", output.Scripts[i], v)
		}
	}
}
