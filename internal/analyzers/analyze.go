package analyzers

import (
	//"encoding/base32"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/RidmaTP/web-analyzer/internal/fetcher"
	"github.com/RidmaTP/web-analyzer/internal/models"
	"github.com/RidmaTP/web-analyzer/internal/utils"
	"golang.org/x/net/html"
)

// body analyzer configuration fields
// stream is the channel used to stream the output to the frontend as a server sent event
// output is the data struct used to define output structure
// muActiveLinks and muInactiveLinks are used to avoid the race conditions for the necessary link slices
// wg is a waitgroup used to synchronize workerpool
// workers define the size of the worker pool
type BodyAnalyzer struct {
	Fetcher         fetcher.BodyFetcher
	Stream          chan string
	Output          models.Output
	muActiveLinks   sync.Mutex
	muInactiveLinks sync.Mutex
	wg              *sync.WaitGroup
	Workers         int
}

// main function of the analyzation process
// gets the reader using fetchbody func
// then it tokenizes the content and goes through the tokens
// all analytics are collected when going through all the tokens once
// during the scraping process, once a result is found they will be pushed to the frontend in realtime using http1.1 SSE
// job queue wth a worker pool is used to improve the performance of finding active/inactive links
func (a *BodyAnalyzer) Analyze(url string) *models.ErrorOut {
	var inTitle bool
	a.muActiveLinks, a.muInactiveLinks = sync.Mutex{}, sync.Mutex{}
	a.wg = &sync.WaitGroup{}
	linkJobQueue := make(chan string, a.Workers)
	loginFlags := models.LoginFlags{}

	ioReader, err := a.Fetcher.FetchBody(url)
	if err != nil {
		return &models.ErrorOut{StatusCode: http.StatusBadGateway, Error: err.Error()}
	}
	defer ioReader.Close()
	tokenizer := html.NewTokenizer(ioReader)

	for i := 0; i < a.Workers; i++ {
		a.wg.Add(1)
		go func(a *BodyAnalyzer, linkJobQueue *chan string, baseUrl string) {
			defer a.wg.Done()
			a.ActiveCheckWorker(baseUrl, linkJobQueue)

		}(a, &linkJobQueue, url)
	}

	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break
			}
			return &models.ErrorOut{StatusCode: http.StatusInternalServerError, Error: err.Error()}
		}
		token := tokenizer.Token()

		isInTitle, err := a.FindTitle(tokenType, token, inTitle)
		if err != nil {
			return &models.ErrorOut{StatusCode: http.StatusInternalServerError, Error: err.Error()}
		}
		inTitle = isInTitle

		err = a.FindHTMLVersion(tokenType, token)
		if err != nil {
			return &models.ErrorOut{StatusCode: http.StatusInternalServerError, Error: err.Error()}
		}

		err = a.FindHeaderCount(tokenType, token)
		if err != nil {
			return &models.ErrorOut{StatusCode: http.StatusInternalServerError, Error: err.Error()}
		}

		err = a.FindLinks(tokenType, token, url, &linkJobQueue)
		if err != nil {
			return &models.ErrorOut{StatusCode: http.StatusInternalServerError, Error: err.Error()}
		}
		err = a.FindIfLogin(tokenType, token, &loginFlags)
		if err != nil {
			return &models.ErrorOut{StatusCode: http.StatusInternalServerError, Error: err.Error()}
		}
	}
	close(linkJobQueue)
	a.wg.Wait()

	return nil
}

// used to find the title of the html body
func (a *BodyAnalyzer) FindTitle(tokenType html.TokenType, token html.Token, inTitle bool) (bool, error) {
	if a.Output.Title != "" {
		return false, nil
	}
	if token.Data == "title" {
		if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
			return true, nil
		} else if tokenType == html.EndTagToken {
			return false, nil
		}
	}
	if tokenType == html.TextToken {
		if inTitle {
			trimmed := strings.TrimSpace(string(token.Data))
			if trimmed != "" {
				a.Output.Title = trimmed
				jsonStr, err := utils.JsonToText(a.Output)
				if err != nil {
					return inTitle, err
				}
				a.Stream <- *jsonStr
				return inTitle, nil
			}
		}
	}
	return inTitle, nil
}

// used to find the version of the html
func (a *BodyAnalyzer) FindHTMLVersion(tokenType html.TokenType, token html.Token) error {
	if a.Output.Version != "" {
		return nil
	}

	version := ""
	if tokenType == html.DoctypeToken {
		doctype := token.Data

		if doctype == "html" {
			version = "HTML5"
		} else if utils.ContainsIgnoreCase(doctype, "XHTML") {
			version = "XHTML"
		} else if utils.ContainsIgnoreCase(doctype, "HTML 4.01") {
			version = "HTML 4.01"
		} else {
			version = doctype
		}
	}
	if version != "" {
		a.Output.Version = version
		jsonStr, err := utils.JsonToText(a.Output)
		if err != nil {
			return err
		}
		a.Stream <- *jsonStr
	}
	return nil
}

// used to find the headercount of the html body
func (a *BodyAnalyzer) FindHeaderCount(tokenType html.TokenType, token html.Token) error {
	if a.Output.Headers == nil {
		a.Output.Headers = make(map[string]int)
	}
	if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
		header := token.Data
		if header == "h1" || header == "h2" || header == "h3" || header == "h4" || header == "h5" || header == "h6" {
			a.Output.Headers[header]++
			jsonStr, err := utils.JsonToText(a.Output)
			if err != nil {
				return err
			}
			a.Stream <- *jsonStr
		}
	}
	return nil
}

// used to find the External,Internal links
// acts as the producer of the linkJobQueue
// when a link is found it checks if its internal/external and then pushes it to the job queue for a worker to check if its available
func (a *BodyAnalyzer) FindLinks(tokenType html.TokenType, token html.Token, baseUrl string, linkJobQueue *chan string) error {
	if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
		tokenData := token.Data
		if tokenData == "a" {
			for _, attr := range token.Attr {
				if attr.Key == "href" {
					if utils.IsExternalLink(attr.Val, baseUrl) {
						a.Output.ExternalLinks.Count++
						a.Output.ExternalLinks.Links = append(a.Output.ExternalLinks.Links, attr.Val)
					} else {
						a.Output.InternalLinks.Count++
						a.Output.InternalLinks.Links = append(a.Output.InternalLinks.Links, attr.Val)
					}
					if linkJobQueue != nil {
						*linkJobQueue <- attr.Val
					}

					jsonStr, err := utils.JsonToText(a.Output)
					if err != nil {
						return err
					}
					a.Stream <- *jsonStr
				}
			}
		}
	}
	return nil
}

// finds if the html body has a login form
func (a *BodyAnalyzer) FindIfLogin(tokenType html.TokenType, token html.Token, loginFlags *models.LoginFlags) error {
	if a.Output.IsLogin {
		return nil
	}
	if loginFlags.IsLoginButton && loginFlags.IsPasswordField && loginFlags.IsTextField && loginFlags.IsForm {
		a.Output.IsLogin = true
		if a.Stream != nil {
			jsonStr, err := utils.JsonToText(a.Output)
			if err != nil {
				return err
			}
			a.Stream <- *jsonStr
		}

		return nil
	}
	if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {

		tokenData := token.Data

		if tokenData == "form" {
			loginFlags.IsForm = true
			loginFlags.InForm = true
			return nil
		} else if tokenData == "input" {

			for _, attr := range token.Attr {
				if attr.Key == "type" {
					if attr.Val == "password" {
						loginFlags.IsPasswordField = true
						return nil
					} else if attr.Val == "email" || attr.Val == "text" {
						loginFlags.IsTextField = true
						return nil
					} else if attr.Val == "submit" {
						loginFlags.IsLoginButton = true
						return nil
					}
				}
			}
		} else if tokenData == "button" {
			for _, attr := range token.Attr {
				if attr.Key == "type" {
					if attr.Val == "submit" {
						loginFlags.InButton = true
						return nil
					}
				}
			}
		}
	} else if tokenType == html.EndTagToken {
		tokenData := token.Data
		if tokenData == "form" {
			return nil

		} else if tokenData == "button" {
			if loginFlags.InForm && loginFlags.InButton {
				loginFlags.InButton = false
				return nil
			}
		}
	} else if tokenType == html.TextToken {
		if loginFlags.InButton && loginFlags.InForm {
			loginKeywords := []string{"login", "log in", "sign in", "signin", "submit", "access"}
			btnText := strings.ToLower(strings.ReplaceAll(token.Data, " ", ""))
			for _, keyword := range loginKeywords {
				if btnText == keyword {
					loginFlags.IsLoginButton = true
					return nil
				}
			}
		}
	}
	return nil
}

// acts as the worker of the job queue
// checks if the link is available/not , groups them and pushes into the data stream as a text obj
func (a *BodyAnalyzer) ActiveCheckWorker(baseUrl string, linkJobQueue *chan string) {
	for link := range *linkJobQueue {
		link = utils.AddInternalHost(link, baseUrl)

		_, err := a.Fetcher.FetchBody(link)
		if err != nil {
			a.muInactiveLinks.Lock()
			a.Output.InactiveLinks.Count++
			a.Output.InactiveLinks.Links = append(a.Output.InactiveLinks.Links, link)
			a.muInactiveLinks.Unlock()
		} else {
			a.muActiveLinks.Lock()
			a.Output.ActiveLinks.Count++
			a.Output.ActiveLinks.Links = append(a.Output.ActiveLinks.Links, link)
			a.muActiveLinks.Unlock()
		}
		jsonStr, err := utils.JsonToText(a.Output)
		if a.Stream != nil {
			a.Stream <- *jsonStr
		}

	}
}
