package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/integralist/go-web-crawler/internal/mapper"
)

// green provides coloured output for text given to a string format function.
var green = color.New(color.FgGreen).SprintFunc()

// map of template functions that enable us to identify the final item within a
// collection being iterated over.
var fns = template.FuncMap{
	"plus1": func(x int) int {
		return x + 1
	},
}

// Dot renders our results in dot format for use with graphviz
func Dot(results []mapper.Page) string {
	dotTmpl := `digraph sitemap { {{- range .}}
  "{{.URL}}"
    -> { {{- $n := len .Anchors}}{{range  $i, $v := .Anchors}}
      "{{.}}"{{if eq (plus1 $i) $n}}{{else}},{{end}}{{end}}
    }{{end}}
}`

	tmpl, err := template.New("digraph").Funcs(fns).Parse(dotTmpl)
	if err != nil {
		log.Fatal(err)
	}

	var output bytes.Buffer
	if err := tmpl.Execute(&output, results); err != nil {
		log.Fatal(err)
	}

	return output.String()
}

// Pretty cleanly formats a given data structure for easily reading.
func Pretty(v interface{}) (s string) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}
	return string(b)
}

// Standard is the default formatted output for the program
func Standard(results []mapper.Page, startTime time.Time) {
	fmt.Printf("-------------------------\n\nNumber of URLs crawled and processed: %s\n", green(len(results)))
	fmt.Printf("Time: %s\n", green(time.Since(startTime)))
}
