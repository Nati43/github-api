package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// NewRequest constructs a new http request
func NewRequest(method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return &http.Request{}, err
	}

	// set request headers
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("accept", "application/json")
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	// basicAuthString := base64.StdEncoding.EncodeToString([]byte(client.ClientId + ":" + client.ClientSecret))
	// req.Header.Set("Authorization", "Basic "+basicAuthString)

	return req, nil
}

// GetNextFromLinkHeader parses the link response header and extracts the link for the next page, if any
func GetNextFromLinkHeader(lh string) string {
	next := ""
	parts := strings.Split(lh, ",")
	for _, part := range parts {
		if strings.Contains(part, "next") {
			next = strings.Split(part, ";")[0]
		}
	}

	// filter out < & >
	next = next[1 : len(next)-1]

	return next
}

// SanitizeRepoURL formulates a proper github api url if not already
func SanitizeRepoURL(url string) (string, error) {
	if strings.Contains(url, "api.github.com") {
		// it already points to the api
		return url, nil
	}

	// split the url and get the path part
	parts := strings.Split(url, "github.com")
	if len(parts) < 2 || parts[1] == "" {
		return "", fmt.Errorf("invalid url provided")
	}
	url = "https://api.github.com/repos" + parts[1]

	return url, nil
}
