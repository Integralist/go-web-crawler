package requester

import (
	"io/ioutil"
	"net/http"

	"github.com/integralist/go-web-crawler/internal/types"
	"github.com/sirupsen/logrus"
)

// HTTPClient is an interface for injecting a preconfigured HTTP client.
type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

// log is a preconfigured logger instance and is set by the coordinator package.
var log *logrus.Entry

// Page represents the requested HTML page (its url & body).
type Page struct {
	URL    string
	Body   []byte
	Status int
}

// Init configures the package from an outside mediator
func Init(instr *types.Instrumentation) {
	log = instr.Logger
}

// Get retrieves the contents of the specified url parameter.
func Get(url string, client HTTPClient) (Page, error) {
	res, err := client.Get(url)
	if err != nil {
		return Page{}, err
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return Page{}, err
	}

	return Page{
		URL:    url,
		Body:   body,
		Status: res.StatusCode,
	}, nil
}
