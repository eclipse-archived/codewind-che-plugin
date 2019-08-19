package che

import (
	"fmt"
	"testing"
)

var CheIngress = "che-eclipse-che.9.1.2.3.nip.io"

func TestParseValidCheAPI(t *testing.T) {
	tests := []struct {
		name   string
		cheAPI string
	}{
		{
			name:   fmt.Sprintf("Parse valid Che API URL, with HTTP"),
			cheAPI: "http://" + CheIngress + "/api",
		},
		{
			name:   fmt.Sprintf("Parse valid Che API URL, with HTTPS"),
			cheAPI: "https://" + CheIngress + "/api",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedIngress, err := GetCheIngress(tt.cheAPI)

			// Test should fail if we're unable to parse a valid Che API URL
			if err != nil {
				t.Error(err)
			}

			// Test should fail if the parsed ingress domain doesn't match our expected ingress domain
			if parsedIngress != CheIngress {
				t.Errorf("Parsed ingress doesn't match expected ingress. Have: %v, expected %v", parsedIngress, CheIngress)
			}
		})
	}
}

func TestParseInvalidCheAPI(t *testing.T) {
	tests := []struct {
		name   string
		cheAPI string
	}{
		{
			name:   fmt.Sprintf("Invalid CHE API URL, no path"),
			cheAPI: "://" + CheIngress,
		},
		{
			name:   fmt.Sprintf("Invalid CHE API URL, with path"),
			cheAPI: "://" + CheIngress + "/api",
		},
		{
			name:   fmt.Sprintf("Invalid Che API URL, no protocol"),
			cheAPI: CheIngress + "/api",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedIngress, err := GetCheIngress(tt.cheAPI)

			// GetCheIngress should return an error here, so the test should fail if it doesn't
			if err == nil || parsedIngress != "" {
				t.Errorf("Che API URL parsing didn't fail as expected. Parsed Che Ingress: %v\n", parsedIngress)
			}
		})
	}
}
