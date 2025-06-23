package analyzers

import (
	"errors"
	"io"
	"strings"

	"github.com/RidmaTP/web-analyzer/internal/fetcher"
	"github.com/RidmaTP/web-analyzer/internal/models"
	"github.com/RidmaTP/web-analyzer/internal/utils"
	"golang.org/x/net/html"
)

type BodyAnalyzer struct {
	Fetcher fetcher.BodyFetcher
	Stream  chan string
	Output  models.Output
}

type LoginFlags struct {
	IsForm          bool
	IsPasswordField bool
	IsTextField     bool
	IsLoginButton   bool
	InForm          bool
	InButton        bool
}

func (a *BodyAnalyzer) Analyze(url string) error {
	ioReader, err := a.Fetcher.FetchBody(url)
	if err != nil {
		return err
	}
	defer ioReader.Close()
	tokenizer := html.NewTokenizer(ioReader)

	var inTitle bool

	loginFlags := LoginFlags{}

	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break
			}
			return errors.New("error tokenzing html : " + err.Error())
		}
		token := tokenizer.Token()

		isInTitle, err := a.FindTitle(tokenType, token, inTitle)
		if err != nil {
			return err
		}
		inTitle = isInTitle

		err = a.FindHTMLVersion(tokenType, token)
		if err != nil {
			return err
		}

		err = a.FindHeaderCount(tokenType, token)
		if err != nil {
			return err
		}

		err = a.FindLinks(tokenType, token, url)
		if err != nil {
			return err
		}
		err = a.FindIfLogin(tokenType, token, &loginFlags)
		if err != nil {
			return err
		}
		//fmt.Println(loginFlags)
	}

	return nil
}

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

func (a *BodyAnalyzer) FindLinks(tokenType html.TokenType, token html.Token, baseUrl string) error {
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
func (a *BodyAnalyzer) FindIfLogin(tokenType html.TokenType, token html.Token, loginFlags *LoginFlags) error {
	if a.Output.IsLogin{
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
