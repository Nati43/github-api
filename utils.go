package main

import (
	"io"
	"net/http"
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
