package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"github.com/RidmaTP/web-analyzer/internal/models"
)

func SendErrResponse(err error) map[string]string {
	return map[string]string{
		"error": err.Error(),
	}
}

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


func ErrStreamObj(errStr string) *string {
	errString := fmt.Sprintf(`{"error" : "%s"}`, errStr)
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

