package analyzers

import (
	"encoding/json"
	//"fmt"
	"testing"

	"github.com/RidmaTP/web-analyzer/internal/fetcher"
	"github.com/RidmaTP/web-analyzer/internal/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func Test_Analyze(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		wantTitle   string
		wantVersion string
		expectErr   bool
	}{
		{
			name: "Basic HTML5 page",
			html: `
				<!DOCTYPE html>
				<html>
					<head>
						<title>Test Page</title>
					</head>
					<body>
						<h1>Welcome</h1>
					</body>
				</html>
			`,
			wantTitle:   "Test Page",
			wantVersion: "HTML5",
			expectErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ba := BodyAnalyzer{
				Fetcher: &fetcher.MockFetcher{
					ResponseBody: tt.html,
				},
				Stream: make(chan string, 10),
				Output: models.Output{},
			}

			err := ba.Analyze("")
			if (err != nil) != tt.expectErr {
				t.Fatalf("Analyze() error = %v, expectErr %v", err, tt.expectErr)
			}

			close(ba.Stream)

			var lastMsg string
			for msg := range ba.Stream {
				lastMsg = msg
			}

			if lastMsg == "" && !tt.expectErr {
				t.Fatal("expected msg from stream, got none")
			}

			var out models.Output
			if lastMsg != "" {
				err = json.Unmarshal([]byte(lastMsg), &out)
				if err != nil {
					t.Fatalf("failed to unmarshal last message: %v", err)
				}

				if out.Title != tt.wantTitle {
					t.Errorf("expected title %q, got %q", tt.wantTitle, out.Title)
				}
				if out.Version != tt.wantVersion {
					t.Errorf("expected version %q, got %q", tt.wantVersion, out.Version)
				}
			}
		})
	}
}

func TestBodyAnalyzer_FindTitle(t *testing.T) {
	tests := []struct {
		name           string
		tokenType      html.TokenType
		token          html.Token
		initialInTitle bool
		wantInTitle    bool
		wantTitle      string
		wantStreamMsg  bool
		wantErr        bool
	}{
		{
			name:           "Start tag <title>",
			tokenType:      html.StartTagToken,
			token:          html.Token{Data: "title"},
			initialInTitle: false,
			wantInTitle:    true,
		},
		{
			name:           "End tag </title>",
			tokenType:      html.EndTagToken,
			token:          html.Token{Data: "title"},
			initialInTitle: true,
			wantInTitle:    false,
		},
		{
			name:           "Text token inside title",
			tokenType:      html.TextToken,
			token:          html.Token{Data: "  My Page Title  "},
			initialInTitle: true,
			wantInTitle:    true,
			wantTitle:      "My Page Title",
			wantStreamMsg:  true,
		},
		{
			name:           "Text token outside title",
			tokenType:      html.TextToken,
			token:          html.Token{Data: "Ignore this"},
			initialInTitle: false,
			wantInTitle:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ba := &BodyAnalyzer{
				Stream: make(chan string, 1),
				Output: models.Output{},
			}

			inTitle, err := ba.FindTitle(tt.tokenType, tt.token, tt.initialInTitle)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantInTitle, inTitle, "inTitle mismatch")

			if tt.wantStreamMsg {
				select {
				case msg := <-ba.Stream:
					var out models.Output
					err := json.Unmarshal([]byte(msg), &out)
					assert.NoError(t, err)
					assert.Equal(t, tt.wantTitle, out.Title)
				default:
					t.Error("expected message on Stream, but none found")
				}
			} else {
				select {
				case msg := <-ba.Stream:
					t.Errorf("expected no message on Stream, but got %q", msg)
				default:
				}
			}
		})
	}
}

func TestBodyAnalyzer_FindHTMLVersion(t *testing.T) {
	tests := []struct {
		name          string
		tokenType     html.TokenType
		token         html.Token
		wantVersion   string
		wantStreamMsg bool
		wantErr       bool
	}{
		{
			name:          "HTML5 doctype",
			tokenType:     html.DoctypeToken,
			token:         html.Token{Data: "html"},
			wantVersion:   "HTML5",
			wantStreamMsg: true,
		},
		{
			name:          "XHTML doctype",
			tokenType:     html.DoctypeToken,
			token:         html.Token{Data: "XHTML 1.0 Strict"},
			wantVersion:   "XHTML",
			wantStreamMsg: true,
		},
		{
			name:          "HTML 4.01 doctype",
			tokenType:     html.DoctypeToken,
			token:         html.Token{Data: "HTML 4.01 Transitional"},
			wantVersion:   "HTML 4.01",
			wantStreamMsg: true,
		},
		{
			name:          "Other doctype",
			tokenType:     html.DoctypeToken,
			token:         html.Token{Data: "Custom Doctype"},
			wantVersion:   "Custom Doctype",
			wantStreamMsg: true,
		},
		{
			name:          "Not a doctype token",
			tokenType:     html.StartTagToken,
			token:         html.Token{Data: "html"},
			wantVersion:   "",
			wantStreamMsg: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ba := &BodyAnalyzer{
				Stream: make(chan string, 1),
				Output: models.Output{},
			}

			err := ba.FindHTMLVersion(tt.tokenType, tt.token)
			assert.NoError(t, err)
			if tt.wantStreamMsg {
				select {
				case msg := <-ba.Stream:
					var out models.Output
					err := json.Unmarshal([]byte(msg), &out)
					assert.NoError(t, err)
					assert.Equal(t, tt.wantVersion, out.Version)
				default:
					t.Error("expected message on Stream, but none found")
				}
			}
		})
	}
}

func TestBodyAnalyzer_FindHeaderCount(t *testing.T) {
	tests := []struct {
		name          string
		tokenType     html.TokenType
		token         html.Token
		initialData   map[string]int
		expected      map[string]int
		wantStreamMsg bool
	}{
		{
			name:          "Counting h1",
			tokenType:     html.StartTagToken,
			token:         html.Token{Data: "h1"},
			expected:      map[string]int{"h1": 1},
			wantStreamMsg: true,
		},
		{
			name:      "Non-header tag",
			tokenType: html.StartTagToken,
			token:     html.Token{Data: "div"},
			expected:  map[string]int{},
		},
		{
			name:      "Non-start tag token",
			tokenType: html.EndTagToken,
			token:     html.Token{Data: "h2"},
			expected:  map[string]int{},
		},
		{
			name:          "Multiple headers count",
			tokenType:     html.StartTagToken,
			token:         html.Token{Data: "h3"},
			initialData:   map[string]int{"h3": 2},
			expected:      map[string]int{"h3": 3},
			wantStreamMsg: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := make(chan string, 1)
			ba := &BodyAnalyzer{
				Output: models.Output{
					Headers: tt.initialData,
				},
				Stream: stream,
			}

			err := ba.FindHeaderCount(tt.tokenType, tt.token)
			assert.NoError(t, err)

			assert.Equal(t, tt.expected, ba.Output.Headers)

			if tt.wantStreamMsg {
				select {
				case msg := <-stream:
					var out models.Output
					err := json.Unmarshal([]byte(msg), &out)
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, ba.Output.Headers)
				default:
					assert.Fail(t, "expected message on Stream, but none found")
				}
			}
		})
	}
}

func TestBodyAnalyzer_FindLinks(t *testing.T) {
	tests := []struct {
		name          string
		tokenType     html.TokenType
		token         html.Token
		expected      models.LinksData
		baseurl       string
		isExternal    bool
		wantStreamMsg bool
	}{
		{
			name:      "Parsing a tag to get external link",
			tokenType: html.StartTagToken,
			token:     html.Token{Data: "a", Attr: []html.Attribute{html.Attribute{Key: "href", Val: "https://www.lucytech.se"}}},
			expected: models.LinksData{
				Count: 1,
				Links: []string{"https://www.lucytech.se"},
			},
			baseurl:       "https://www.home24.de",
			wantStreamMsg: true,
			isExternal:    true,
		},
		{
			name:      "Parsing a tag to get internal link",
			tokenType: html.StartTagToken,
			token:     html.Token{Data: "a", Attr: []html.Attribute{html.Attribute{Key: "href", Val: "https://www.lucytech.se"}}},
			expected: models.LinksData{
				Count: 1,
				Links: []string{"https://www.lucytech.se"},
			},
			baseurl:       "https://www.lucytech.se",
			wantStreamMsg: true,
			isExternal:    false,
		},
		{
			name:      "Ignoring non a tags",
			tokenType: html.StartTagToken,
			token:     html.Token{Data: "div"},
			expected:  models.LinksData{},
		},
		{
			name:      "Ignoring non href attributes",
			tokenType: html.StartTagToken,
			token:     html.Token{Data: "a", Attr: []html.Attribute{html.Attribute{Key: "type", Val: "text/html"}}},
			expected:  models.LinksData{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := make(chan string, 1)
			ba := &BodyAnalyzer{
				Output: models.Output{},
				Stream: stream,
			}

			err := ba.FindLinks(tt.tokenType, tt.token, tt.baseurl)
			assert.NoError(t, err)
			if tt.isExternal {
				assert.Equal(t, tt.expected, ba.Output.ExternalLinks)
			} else {
				assert.Equal(t, tt.expected, ba.Output.InternalLinks)
			}

			if tt.wantStreamMsg {
				select {
				case msg := <-stream:
					var out models.Output
					err := json.Unmarshal([]byte(msg), &out)
					assert.NoError(t, err)
					if tt.isExternal {
						assert.Equal(t, tt.expected, ba.Output.ExternalLinks)
					} else {
						assert.Equal(t, tt.expected, ba.Output.InternalLinks)
					}
				default:
					assert.Fail(t, "expected message on Stream, but none found")
				}
			}
		})
	}
}
