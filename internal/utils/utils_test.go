package utils

import (
	"encoding/json"
	"testing"

	"github.com/RidmaTP/web-analyzer/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestContainsIgnoreCase(t *testing.T) {
	testCases := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"exact match", "Lucytech", "Lucytech", true},
		{"case insensitive", "Lucytech", "lucytech", true},
		{"partial match", "Lucytech", "Lucy", true},
		{"no match", "Lucytech", "home24", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := ContainsIgnoreCase(tc.s, tc.substr)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestJsonToText(t *testing.T) {
	testCases := []struct {
		name       string
		input      models.Output
		expectErr bool
	}{
		{
			name:       "valid output",
			input:      models.Output{Title: "My Page", Version: "HTML5"},
			expectErr: false,
		},
		{
			name:       "empty struct",
			input:      models.Output{},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			str, err := JsonToText(tc.input)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, str)
				var decoded models.Output
				assert.NoError(t, json.Unmarshal([]byte(*str), &decoded))
			}
		})
	}
}
