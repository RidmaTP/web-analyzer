package analyzers

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/RidmaTP/web-analyzer/internal/fetcher"
	"github.com/RidmaTP/web-analyzer/internal/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func Test_Analyze(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		expectTitle   string
		expectVersion string
		expectErr     bool
		forceErr      bool
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
			expectTitle:   "Test Page",
			expectVersion: "HTML5",
			expectErr:     false,
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
				erro := json.Unmarshal([]byte(lastMsg), &out)
				if erro != nil {
					t.Fatalf("failed to unmarshal last message: %v", err)
				}

				if out.Title != tt.expectTitle {
					t.Errorf("expected title %q, got %q", tt.expectTitle, out.Title)
				}
				if out.Version != tt.expectVersion {
					t.Errorf("expected version %q, got %q", tt.expectVersion, out.Version)
				}
			}
		})
	}
}
func Test_Analyze_Err(t *testing.T) {
	tests := []struct {
		name            string
		html            string
		expectErr       bool
		forceErrFetcher bool
		forceErrReader  bool
	}{
		{
			name: "Forcing Fetcher error",
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
			expectErr:       true,
			forceErrFetcher: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ba := BodyAnalyzer{
				Fetcher: &fetcher.MockFetcher{
					ResponseBody:   tt.html,
					ForceErr:       tt.forceErrFetcher,
					ForceReaderErr: tt.forceErrReader,
				},
				Stream: make(chan string, 10),
				Output: models.Output{},
			}

			err := ba.Analyze("")
			assert.Equal(t, tt.expectErr, err != nil)

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
				erro := json.Unmarshal([]byte(lastMsg), &out)
				if erro != nil {
					t.Fatalf("failed to unmarshal last message: %v", err)
				}
			}
		})
	}
}

func Test_FindTitle(t *testing.T) {
	tests := []struct {
		name          string
		tokenType     html.TokenType
		token         html.Token
		initInTitle   bool
		inTitle       bool
		isTitle       string
		isStream      bool
		wantErr       bool
		existingTitle string
	}{
		{
			name:        "title as start tag",
			tokenType:   html.StartTagToken,
			token:       html.Token{Data: "title"},
			initInTitle: false,
			inTitle:     true,
		},
		{
			name:        "title as end tag",
			tokenType:   html.EndTagToken,
			token:       html.Token{Data: "title"},
			initInTitle: true,
			inTitle:     false,
		},
		{
			name:        "Text token inside title tags",
			tokenType:   html.TextToken,
			token:       html.Token{Data: "  My Page Title  "},
			initInTitle: true,
			inTitle:     true,
			isTitle:     "My Page Title",
			isStream:    true,
		},
		{
			name:        "Text token outside title tags",
			tokenType:   html.TextToken,
			token:       html.Token{Data: "Ignore this"},
			initInTitle: false,
			inTitle:     false,
		},
		{
			name:          "Title Already found",
			tokenType:     html.TextToken,
			token:         html.Token{Data: "New title"},
			initInTitle:   true,
			inTitle:       false,
			isStream:      false,
			existingTitle: "Old title",
			isTitle:       "Old title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ba := &BodyAnalyzer{
				Stream: make(chan string, 1),
				Output: models.Output{Title: tt.existingTitle},
			}

			inTitle, err := ba.FindTitle(tt.tokenType, tt.token, tt.initInTitle)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.inTitle, inTitle, "inTitle mismatch")

			if tt.isStream {
				select {
				case msg := <-ba.Stream:
					var out models.Output
					err := json.Unmarshal([]byte(msg), &out)
					assert.NoError(t, err)
					assert.Equal(t, tt.isTitle, out.Title, "expected "+tt.existingTitle+", got "+out.Title)
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

func Test_FindHTMLVersion(t *testing.T) {
	tests := []struct {
		name          string
		tokenType     html.TokenType
		token         html.Token
		expectVersion string
		isStream      bool
		expectErr     bool
	}{
		{
			name:          "HTML5 doctype",
			tokenType:     html.DoctypeToken,
			token:         html.Token{Data: "html"},
			expectVersion: "HTML5",
			isStream:      true,
		},
		{
			name:          "XHTML doctype",
			tokenType:     html.DoctypeToken,
			token:         html.Token{Data: "XHTML 1.0 Strict"},
			expectVersion: "XHTML",
			isStream:      true,
		},
		{
			name:          "HTML 4.01 doctype",
			tokenType:     html.DoctypeToken,
			token:         html.Token{Data: "HTML 4.01 Transitional"},
			expectVersion: "HTML 4.01",
			isStream:      true,
		},
		{
			name:          "Other doctype",
			tokenType:     html.DoctypeToken,
			token:         html.Token{Data: "Custom Doctype"},
			expectVersion: "Custom Doctype",
			isStream:      true,
		},
		{
			name:          "Not a doctype token",
			tokenType:     html.StartTagToken,
			token:         html.Token{Data: "html"},
			expectVersion: "",
			isStream:      false,
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
			if tt.isStream {
				select {
				case msg := <-ba.Stream:
					var out models.Output
					err := json.Unmarshal([]byte(msg), &out)
					assert.NoError(t, err)
					assert.Equal(t, tt.expectVersion, out.Version)
				default:
					t.Error("expected message on Stream, but none found")
				}
			}
		})
	}
}

func Test_FindHeaderCount(t *testing.T) {
	tests := []struct {
		name        string
		tokenType   html.TokenType
		token       html.Token
		initialData map[string]int
		expected    map[string]int
		isStream    bool
	}{
		{
			name:      "Counting h1",
			tokenType: html.StartTagToken,
			token:     html.Token{Data: "h1"},
			expected:  map[string]int{"h1": 1},
			isStream:  true,
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
			name:        "Multiple headers count",
			tokenType:   html.StartTagToken,
			token:       html.Token{Data: "h3"},
			initialData: map[string]int{"h3": 2},
			expected:    map[string]int{"h3": 3},
			isStream:    true,
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

			if tt.isStream {
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

func Test_FindLinks(t *testing.T) {
	tests := []struct {
		name       string
		tokenType  html.TokenType
		token      html.Token
		expected   models.LinksData
		baseurl    string
		isExternal bool
		isStream   bool
	}{
		{
			name:      "Parsing a tag to get external link",
			tokenType: html.StartTagToken,
			token:     html.Token{Data: "a", Attr: []html.Attribute{html.Attribute{Key: "href", Val: "https://www.lucytech.se"}}},
			expected: models.LinksData{
				Count: 1,
				Links: []string{"https://www.lucytech.se"},
			},
			baseurl:    "https://www.home24.de",
			isStream:   true,
			isExternal: true,
		},
		{
			name:      "Parsing a tag to get external link (link tag)",
			tokenType: html.StartTagToken,
			token:     html.Token{Data: "link", Attr: []html.Attribute{html.Attribute{Key: "href", Val: "https://www.lucytech.se"}}},
			expected: models.LinksData{
				Count: 1,
				Links: []string{"https://www.lucytech.se"},
			},
			baseurl:    "https://www.home24.de",
			isStream:   true,
			isExternal: true,
		},
		{
			name:      "Parsing a tag to get internal link",
			tokenType: html.StartTagToken,
			token:     html.Token{Data: "a", Attr: []html.Attribute{html.Attribute{Key: "href", Val: "https://www.lucytech.se"}}},
			expected: models.LinksData{
				Count: 1,
				Links: []string{"https://www.lucytech.se"},
			},
			baseurl:    "https://www.lucytech.se",
			isStream:   true,
			isExternal: false,
		},
		{
			name:      "Parsing a tag to get internal link (link tag)",
			tokenType: html.StartTagToken,
			token:     html.Token{Data: "link", Attr: []html.Attribute{html.Attribute{Key: "href", Val: "https://www.lucytech.se"}}},
			expected: models.LinksData{
				Count: 1,
				Links: []string{"https://www.lucytech.se"},
			},
			baseurl:    "https://www.lucytech.se",
			isStream:   true,
			isExternal: false,
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
				Output:          models.Output{},
				Stream:          stream,
				Fetcher:         &fetcher.MockFetcher{},
				muInactiveLinks: sync.Mutex{},
				muActiveLinks:   sync.Mutex{},
				wg:              &sync.WaitGroup{},
			}
			jobs := make(chan string, 10)
			err := ba.FindLinks(tt.tokenType, tt.token, tt.baseurl, &jobs)
			assert.NoError(t, err)
			if tt.isExternal {
				assert.Equal(t, tt.expected, ba.Output.ExternalLinks)
			} else {
				assert.Equal(t, tt.expected, ba.Output.InternalLinks)
			}

			if tt.isStream {
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

func Test_FindIfLogin(t *testing.T) {
	type step struct {
		tokenType html.TokenType
		token     html.Token
	}

	tests := []struct {
		name              string
		steps             []step
		expectFlags       models.LoginFlags
		expectIsLogin     bool
		alreadyFoundLogin bool
	}{
		{
			name:        "Detect full login with input submit password text",
			expectFlags: models.LoginFlags{IsForm: true, IsPasswordField: true, IsTextField: true, IsLoginButton: true, InForm: true, InButton: false},
			steps: []step{
				{html.StartTagToken, html.Token{Data: "form"}},
				{html.SelfClosingTagToken, html.Token{Data: "input", Attr: []html.Attribute{{Key: "type", Val: "text"}}}},
				{html.SelfClosingTagToken, html.Token{Data: "input", Attr: []html.Attribute{{Key: "type", Val: "password"}}}},
				{html.SelfClosingTagToken, html.Token{Data: "input", Attr: []html.Attribute{{Key: "type", Val: "submit"}}}},
				{html.EndTagToken, html.Token{Data: "form"}},
			},
			expectIsLogin: true,
		},
		{
			name:        "No password field",
			expectFlags: models.LoginFlags{IsForm: true, IsPasswordField: false, IsTextField: true, IsLoginButton: true, InForm: true, InButton: false},
			steps: []step{
				{html.StartTagToken, html.Token{Data: "form"}},
				{html.SelfClosingTagToken, html.Token{Data: "input", Attr: []html.Attribute{{Key: "type", Val: "text"}}}},
				//{html.SelfClosingTagToken, html.Token{Data: "input", Attr: []html.Attribute{{Key: "type", Val: "password"}}}},
				{html.SelfClosingTagToken, html.Token{Data: "input", Attr: []html.Attribute{{Key: "type", Val: "submit"}}}},
				{html.EndTagToken, html.Token{Data: "form"}},
			},
			expectIsLogin: false,
		},
		{
			name:        "Button login text detection",
			expectFlags: models.LoginFlags{IsForm: true, IsPasswordField: false, IsTextField: false, IsLoginButton: true, InForm: true, InButton: false},
			steps: []step{
				{html.StartTagToken, html.Token{Data: "form"}},
				{html.StartTagToken, html.Token{Data: "button", Attr: []html.Attribute{{Key: "type", Val: "submit"}}}},
				{html.TextToken, html.Token{Data: "login"}},
				{html.EndTagToken, html.Token{Data: "button"}},
				{html.EndTagToken, html.Token{Data: "form"}},
			},
			expectIsLogin: false,
		},
		{
			name:        "Unrelated token",
			expectFlags: models.LoginFlags{IsForm: false, IsPasswordField: false, IsTextField: false, IsLoginButton: false, InForm: false, InButton: false},
			steps: []step{
				{html.StartTagToken, html.Token{Data: "h1"}},
				{html.TextToken, html.Token{Data: "Header"}},
				{html.EndTagToken, html.Token{Data: "h1"}},
			},
			expectIsLogin: false,
		},
		{
			name:        "Login Already Found",
			expectFlags: models.LoginFlags{IsForm: false, IsPasswordField: false, IsTextField: false, IsLoginButton: false, InForm: false, InButton: false},
			steps: []step{
				{html.StartTagToken, html.Token{Data: "h1"}},
			},
			expectIsLogin:     true,
			alreadyFoundLogin: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := &BodyAnalyzer{
				Output: models.Output{IsLogin: tt.alreadyFoundLogin},
			}
			flags := models.LoginFlags{}

			for _, step := range tt.steps {
				err := analyzer.FindIfLogin(step.tokenType, step.token, &flags)
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectFlags, flags)
			assert.Equal(t, tt.expectIsLogin, analyzer.Output.IsLogin)
		})
	}
}
func Test_ActiveCheckerWorker(t *testing.T) {
	type step struct {
		tokenType html.TokenType
		token     html.Token
	}

	tests := []struct {
		name          string
		url           string
		jobQueue      chan string
		expected      models.Output
		isUrlInactive bool
	}{
		{
			name:          "Accessible URL",
			url:           "https://lucytech.se/",
			jobQueue:      make(chan string, 1),
			expected:      models.Output{ActiveLinks: models.LinksData{Count: 1, Links: []string{"https://lucytech.se/"}}},
			isUrlInactive: false,
		},
		{
			name:          "Inaccessible URL",
			url:           "https://lucytech.se/",
			jobQueue:      make(chan string, 1),
			expected:      models.Output{InactiveLinks: models.LinksData{Count: 1, Links: []string{"https://lucytech.se/"}}},
			isUrlInactive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := &BodyAnalyzer{
				Output:  models.Output{},
				Fetcher: &fetcher.MockFetcher{ForceErr: tt.isUrlInactive},
				Stream:  make(chan string, 1),
			}

			go analyzer.ActiveCheckWorker(tt.url, &tt.jobQueue)

			tt.jobQueue <- tt.url
			close(tt.jobQueue)
			var out models.Output
			msg := <-analyzer.Stream
			err := json.Unmarshal([]byte(msg), &out)
			assert.NoError(t, err)

			assert.Equal(t, tt.expected, out)
		})
	}
}
