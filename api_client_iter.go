//go:build go1.23

package flareio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/url"
)

type IterResult struct {
	Response *http.Response
	Next     string
}

func createPagingIterator(
	fetchPage func(from string) (*http.Response, error),
) iter.Seq2[*IterResult, error] {
	type ResponseWithNext struct {
		Next string `json:"next"`
	}

	cursor := ""

	return func(yield func(*IterResult, error) bool) {
		for {
			// Fire the request
			response, err := fetchPage(cursor)
			if err != nil {
				yield(nil, err)
				return
			}

			// Read the body
			body, err := io.ReadAll(response.Body)
			if err != nil {
				yield(
					nil,
					fmt.Errorf("failed to read response: %w", err),
				)
				return
			}

			// Close the body
			if err := response.Body.Close(); err != nil {
				yield(
					nil,
					fmt.Errorf("failed to close http response: %w", err),
				)
				return
			}

			// Look for the next token
			var responseWithNext ResponseWithNext
			if err := json.Unmarshal(body, &responseWithNext); err != nil {
				yield(
					nil,
					fmt.Errorf("failed to unmarshal response: %w", err),
				)
				return
			}

			// Replace the body and return the response
			response.Body = io.NopCloser(bytes.NewReader(body))
			if !yield(
				&IterResult{
					Response: response,
					Next:     responseWithNext.Next,
				},
				nil,
			) {
				return
			}

			// Was this the last page?
			if responseWithNext.Next == "" {
				return
			}

			// Set the cursor for next page
			cursor = responseWithNext.Next
		}
	}
}

// IterGet allows to iterate over responses for an API endpoint that
// supports the Flare standard paging pattern.
func (client *ApiClient) IterGet(
	path string,
	params *url.Values,
) iter.Seq2[*IterResult, error] {
	return createPagingIterator(
		func(cursor string) (*http.Response, error) {
			if cursor != "" {
				if params == nil {
					params = &url.Values{}
				}
				params.Set("from", cursor)
			}
			return client.Get(
				path,
				params,
			)
		},
	)
}

// IterPostJson allows to iterate over responses for an API endpoint that
// supports the Flare standard paging pattern.
func (client *ApiClient) IterPostJson(
	path string,
	params *url.Values,
	body map[string]interface{},
) iter.Seq2[*IterResult, error] {
	return createPagingIterator(
		func(cursor string) (*http.Response, error) {
			if cursor != "" {
				body["from"] = cursor
			}

			encodedJson, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body to JSON: %w", err)
			}

			return client.Post(
				path,
				params,
				"application/json",
				bytes.NewReader(encodedJson),
			)
		},
	)

}
