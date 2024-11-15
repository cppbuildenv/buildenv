package io

import (
	"fmt"
	"net/http"
	"time"
)

func CheckAvailable(url string) error {
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	// Check URL availability using HEAD request.
	resp, err := client.Head(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if the status code is in the 2xx range.
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return nil
	}

	return fmt.Errorf("status code: %d", resp.StatusCode)
}
