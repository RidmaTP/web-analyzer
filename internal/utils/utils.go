package utils

import (
	"encoding/json"
	"errors"
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

func JsonToText(output models.Output) (*string, error){
	jsonBytes, err := json.Marshal(output)
	if err != nil {
		return nil, errors.New("err when marshelling : "+err.Error())
	}
	str := string(jsonBytes)
	return &str, nil
}
