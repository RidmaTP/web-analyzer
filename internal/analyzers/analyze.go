package analyzers

import (
	"errors"
	"fmt"
	"io"

	"github.com/RidmaTP/web-analyzer/internal/fetcher"
	"github.com/RidmaTP/web-analyzer/internal/models"
	"github.com/RidmaTP/web-analyzer/internal/utils"
	"golang.org/x/net/html"
)

func Analyze(url string, f fetcher.BodyFetcher) error {
	ioReader, err := f.FetchBody(url)
	if err != nil {
		return err
	}
	defer ioReader.Close()
	tokenizer := html.NewTokenizer(ioReader)

	var inTitle bool
	output := models.Output{}

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
		if output.Title == "" {
			isInTitle, text := FindTitle(tokenType, token, inTitle)
			inTitle = isInTitle
			output.Title = text
		}
		if output.Version == "" {
			output.Version = FindHTMLVersion(tokenType, token)
		}

	}

	fmt.Println("Page title:", output.Title)
	fmt.Println("Page Version:", output.Version)
	return nil
}

func FindTitle(tokenType html.TokenType, token html.Token, inTitle bool) (bool, string) {
	if token.Data == "title" {
		if tokenType == html.StartTagToken {
			return true, ""
		} else if tokenType == html.EndTagToken {
			return false, ""
		}
	}
	if tokenType == html.TextToken {
		if inTitle {
			return inTitle, string(token.Data)
		}
	}
	return inTitle, ""
}

func FindHTMLVersion(tokenType html.TokenType, token html.Token) string {
	if tokenType == html.DoctypeToken {
		doctype := token.Data

		if doctype == "html" {
			return "HTML5"
		} else if utils.ContainsIgnoreCase(doctype, "XHTML") {
			return "XHTML"
		} else if utils.ContainsIgnoreCase(doctype, "HTML 4.01") {
			return "HTML 4.01"
		} else {
			return doctype
		}
	}
	return ""
}