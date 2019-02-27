package requester

import (
	"io/ioutil"
	"net/http"
)

// HTTPClient is an interface for injecting a preconfigured HTTP client.
type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

// Page represents the requested HTML page (its url & body).
type Page struct {
	URL    string
	Body   []byte
	Status int
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
