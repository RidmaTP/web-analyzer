package utils

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/RidmaTP/web-analyzer/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestContainsIgnoreCase(t *testing.T) {
	testCases := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"exact match", "Lucytech", "Lucytech", true},
		{"case insensitive", "Lucytech", "lucytech", true},
		{"partial match", "Lucytech", "Lucy", true},
		{"no match", "Lucytech", "home24", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := ContainsIgnoreCase(tc.s, tc.substr)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestJsonToText(t *testing.T) {
	testCases := []struct {
		name      string
		input     models.Output
		expectErr bool
	}{
		{
			name:      "valid output",
			input:     models.Output{Title: "My Page", Version: "HTML5"},
			expectErr: false,
		},
		{
			name:      "empty struct",
			input:     models.Output{},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			str, err := JsonToText(tc.input)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, str)
				var decoded models.Output
				assert.NoError(t, json.Unmarshal([]byte(*str), &decoded))
			}
		})
	}
}
func TestErrStreamObj(t *testing.T) {
	testCases := []struct {
		name      string
		input     models.ErrorOut
		expectOut string
	}{
		{
			name:      "valid output",
			input:     models.ErrorOut{StatusCode: 400 , Error: "bad request"},
			expectOut: fmt.Sprintf(`{"error" : "%s", "status_code" : "%d"}`, "bad request", 400),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			str := ErrStreamObj(tc.input)
			assert.Equal(t, tc.expectOut, *str)
		})
	}
}
func TestIsExternalLink(t *testing.T) {
	testCases := []struct {
		name    string
		baseurl string
		link    string
		expect  bool
	}{
		{
			name:    "Internal with host",
			baseurl: "https://lucytech.se/",
			link:    "https://lucytech.se/contact",
			expect:  false,
		},
		{
			name:    "Internal without host",
			baseurl: "https://lucytech.se/",
			link:    "/contact",
			expect:  false,
		},
		{
			name:    "External",
			baseurl: "https://lucytech.se/",
			link:    "https://www.home24.de/",
			expect:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := IsExternalLink(tc.link, tc.baseurl)
			assert.Equal(t, r, tc.expect)
		})
	}
}
func TestAddInternalHost(t *testing.T) {
	testCases := []struct {
		name    string
		baseurl string
		link    string
		expect  string
	}{
		{
			name:    "Internal without host",
			baseurl: "https://lucytech.se/",
			link:    "/contact",
			expect:  "https://lucytech.se/contact",
		},
		{
			name:    "Internal without host",
			baseurl: "https://lucytech.se/",
			link:    "https://lucytech.se/contact",
			expect:  "https://lucytech.se/contact",
		},
		{
			name:    "External with host",
			baseurl: "https://lucytech.se/",
			link:    "https://www.home24.de",
			expect:  "https://www.home24.de",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := AddInternalHost(tc.link, tc.baseurl)
			assert.Equal(t, r, tc.expect)
		})
	}
}

func TestValidateUrl(t *testing.T) {
	testCases := []struct {
		name    string
		baseurl string
		expect  *models.ErrorOut
	}{
		{
			name:    "Valid Url",
			baseurl: "https://lucytech.se/",
			expect:  nil,
		},
		{
			name:    "Valid Url",
			baseurl: "https://www.lucytech.se/",
			expect:  nil,
		},
		{
			name:    "missing scheme",
			baseurl: "htt://lucytech.se/",
			expect:  &models.ErrorOut{StatusCode: 400, Error: "url scheme not found"},
		},
		{
			name:    "missing domain without www.",
			baseurl: "http://hello",
			expect:  &models.ErrorOut{StatusCode: 400, Error: "url domain not found"},
		},
		{
			name:    "missing domain with www.",
			baseurl: "http://www.hello",
			expect:  &models.ErrorOut{StatusCode: 400, Error: "url domain not found"},
		},
		{
			name:    "missing scheme completely",
			baseurl: "www.hello.com",
			expect:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := UrlValidationCheck(&tc.baseurl)
			assert.Equal(t, r, tc.expect)
		})
	}
}
