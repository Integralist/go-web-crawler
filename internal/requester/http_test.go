package requester

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

type MockHTTPClient struct{}

func (mhc *MockHTTPClient) Get(url string) (*http.Response, error) {
	body := "foobar"

	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
	}, nil
}

func TestGet(t *testing.T) {
	input := "http://www.foo.com/bar"

	output := Page{
		URL:    input,
		Body:   []byte("foobar"),
		Status: 200,
	}

	mockHTTPclient := MockHTTPClient{}

	actual, _ := Get(input, &mockHTTPclient)

	if actual.URL != output.URL {
		t.Errorf("expected: %+v\ngot: %+v", output.URL, actual.URL)
	}

	if actual.Status != output.Status {
		t.Errorf("expected: %+v\ngot: %+v", output.Status, actual.Status)
	}

	stringActualBody := string(actual.Body)
	stringOutputBody := string(output.Body)

	if stringActualBody != stringOutputBody {
		t.Errorf("expected: %+v\ngot: %+v", stringOutputBody, stringActualBody)
	}
}
