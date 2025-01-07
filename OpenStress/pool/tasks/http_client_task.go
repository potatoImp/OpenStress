package tasks

import (
	"fmt"
	"net/http"
	"io/ioutil"
)

// HttpClient struct definition
type HttpClient struct {
	Name string
	URL  string
}

// Execute executes HTTP request
func (h *HttpClient) Execute() error {
	resp, err := http.Get(h.URL)
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	fmt.Printf("Response: %s\n", body)
	return nil
}
