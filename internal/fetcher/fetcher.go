package fetcher

import (
	"fmt"
	"io"
	"net/http"
)
// fetcher contract
type BodyFetcher interface {
	FetchBody(url string) (io.ReadCloser, error)
}

type Fetcher struct {
}

// Returns the reader to read the body
func (f *Fetcher) FetchBody(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d is returned",resp.StatusCode)
	}
	return resp.Body, nil
}
