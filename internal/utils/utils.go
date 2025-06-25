package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/RidmaTP/web-analyzer/internal/models"
)

// utility functions are included here

func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func JsonToText(output models.Output) (*string, error) {
	jsonBytes, err := json.Marshal(output)
	if err != nil {
		return nil, errors.New("err when marshelling : " + err.Error())
	}
	str := string(jsonBytes)
	return &str, nil
}

func ErrStreamObj(err models.ErrorOut) *string {
	errString := fmt.Sprintf(`{"error" : "%s", "status_code" : "%d"}`, err.Error, err.StatusCode)
	return &errString
}

func IsExternalLink(link, baseUrl string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	bu, err := url.Parse(baseUrl)
	if err != nil {
		return false
	}
	if u.Host == "" {
		return false
	}
	return !strings.Contains(u.Host, bu.Host)
}

func AddInternalHost(link, baseUrl string) string {
	u, _ := url.Parse(link)
	bu, _ := url.Parse(baseUrl)

	if u.Host == "" {
		return bu.Scheme + "://" + bu.Host + link
	}
	return link
}

func UrlValidationCheck(input string) *models.ErrorOut {
	errOut := models.ErrorOut{StatusCode: http.StatusBadRequest, Error: "invalid url"}
	parsedVal, err := url.ParseRequestURI(input)
	if err != nil {
		return &errOut
	}
	if parsedVal.Scheme != "http" && parsedVal.Scheme != "https" {
		errOut.Error = "url scheme not found"
		return &errOut
	}
	hostStr := ""
	if strings.Contains(parsedVal.Host, "www.") {
		wwwRemoved := strings.Split(parsedVal.Host, "www.")[1]
		hostStr = wwwRemoved
	} else {
		hostStr = parsedVal.Host
	}
	if !strings.Contains(hostStr, ".") {
		errOut.Error = "url domain not found"
		return &errOut
	}

	if parsedVal.Host == "" {
		errOut.Error = "url host not found"
		return &errOut
	}

	return nil
}
