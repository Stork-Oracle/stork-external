package publisher_agent

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func RestQuery(method string, baseUrl string, query url.Values, requestBody io.Reader, header http.Header) ([]byte, error) {
	// Parse the base URL
	parsedURL, err := url.Parse(baseUrl)
	if err != nil {
		return nil, fmt.Errorf("error parsing url %s: %v", baseUrl, err)
	}

	if query != nil {
		parsedURL.RawQuery = query.Encode()
	}

	urlString := parsedURL.String()
	req, err := http.NewRequest(method, urlString, requestBody)
	req.Header = header

	if err != nil {
		return nil, fmt.Errorf("error creating %s request for url %s: %v", method, urlString, err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making %s request for url %s: %v", method, urlString, err)

	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response responseBody for %s request for url %s: %v", method, urlString, err)
	}

	return responseBody, nil
}
