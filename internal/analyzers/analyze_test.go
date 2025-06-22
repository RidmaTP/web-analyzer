package analyzers

import (
	"encoding/json"
	"testing"

	"github.com/RidmaTP/web-analyzer/internal/fetcher"
	"github.com/RidmaTP/web-analyzer/internal/models"
	"golang.org/x/net/html"
	"github.com/stretchr/testify/assert"
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
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

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