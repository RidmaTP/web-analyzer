package fetcher

import (
	"errors"
	"io"
	"strings"
)
// This is used to mock the FetchBody using Fetcher interface
// can force errors to test fetcher errors

type MockFetcher struct {
	ResponseBody   string
	ForceErr       bool
	ForceReaderErr bool
}

func (f *MockFetcher) FetchBody(url string) (io.ReadCloser, error) {
	if f.ForceErr {
		return nil, errors.New("mock err")
	}
	if f.ForceReaderErr {
		return &ErrorReader{}, nil
	}
	return io.NopCloser(strings.NewReader(f.ResponseBody)), nil
}

type ErrorReader struct{}

func (e *ErrorReader) Read(p []byte) (int, error) {
	return 0, errors.New("simulated read error")
}

func (e *ErrorReader) Close() error {
	return nil
}