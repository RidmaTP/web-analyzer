package fetcher

import (
	"errors"
	"io"
	"net/http"
	"strings"
)

type BodyFetcher interface {
	FetchBody(url string) (io.ReadCloser, error)
}

type Fetcher struct {
}

func (f *Fetcher) FetchBody(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(resp))
	}
	return resp.Body, nil
}

type MockFetcher struct {
	ResponseBody string
}

func (f *MockFetcher) FetchBody(url string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(f.ResponseBody)), nil
}
