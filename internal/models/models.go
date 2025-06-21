package models

type Result struct {
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
