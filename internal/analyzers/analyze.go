package analyzers

import (
	"errors"
	"fmt"
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

func (a *BodyAnalyzer) Analyze(url string) error {
	ioReader, err := a.Fetcher.FetchBody(url)
	if err != nil {
		return err
	}
	defer ioReader.Close()
	tokenizer := html.NewTokenizer(ioReader)

	var inTitle bool

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

	}

	fmt.Println("Page title:", a.Output.Title)
	fmt.Println("Page Version:", a.Output.Version)
	return nil
}

func (a *BodyAnalyzer) FindTitle(tokenType html.TokenType, token html.Token, inTitle bool) (bool, error) {
	if token.Data == "title" {
		if tokenType == html.StartTagToken {
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
					return inTitle, nil
				}
				a.Stream <- *jsonStr
				return inTitle, nil
			}
		}
	}
	return inTitle, nil
}

func (a *BodyAnalyzer) FindHTMLVersion(tokenType html.TokenType, token html.Token) error {
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
