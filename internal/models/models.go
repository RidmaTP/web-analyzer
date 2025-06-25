package models

// all the data models will be listed here (since there are few models, included everything in one file)
type Output struct {
	Version       string
	Title         string
	Headers       map[string]int
	InternalLinks LinksData
	ExternalLinks LinksData
	ActiveLinks   LinksData
	InactiveLinks LinksData
	IsLogin       bool
}

type LinksData struct {
	Count int
	Links []string
}

type Input struct {
	Url string `json:"url"`
}

type LoginFlags struct {
	IsForm          bool
	IsPasswordField bool
	IsTextField     bool
	IsLoginButton   bool
	InForm          bool
	InButton        bool
}

type ErrorOut struct {
	StatusCode int
	Error      string
}
