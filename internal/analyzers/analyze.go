package analyzers

import (
	"errors"
	"fmt"
	"io"

	"github.com/RidmaTP/web-analyzer/internal/fetcher"
	"github.com/RidmaTP/web-analyzer/internal/models"
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

	}

	fmt.Println("Page title:", output.Title)
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
