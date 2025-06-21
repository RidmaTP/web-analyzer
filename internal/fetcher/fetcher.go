package fetcher

import (
	"errors"
	"io"
	"net/http"
)

func FetchBody(url string) (io.ReadCloser, error) {
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
