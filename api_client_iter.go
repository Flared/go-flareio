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

// IterResult contains results for a given page.
type IterResult struct {
	// Response associated with the fetched page.
	//
	// The response's body must be closed.
	Response *http.Response

	// Next is the token to be used to fetch the next page.
	Next string
}

func getIterResult(
	fetchPage func(from string) (*http.Response, error),
	cursor string,
) (*IterResult, error) {
	response, err := fetchPage(cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch next page: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"got http status code %d status while fetching next page: %s",
			response.StatusCode,
			body,
		)
	}

	type ResponseWithNext struct {
		Next string `json:"next"`
	}
	var responseWithNext ResponseWithNext
	if err := json.Unmarshal(body, &responseWithNext); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	response.Body = io.NopCloser(bytes.NewReader(body))

	iterResult := &IterResult{
		Response: response,
		Next:     responseWithNext.Next,
	}

	return iterResult, nil
}

func createPagingIterator(
	fetchPage func(from string) (*http.Response, error),
) iter.Seq2[*IterResult, error] {
	cursor := ""
	return func(yield func(*IterResult, error) bool) {
		for {
			iterResult, err := getIterResult(
				fetchPage,
				cursor,
			)
			if err != nil {
				yield(nil, err)
				return
			}
			cursor = iterResult.Next
			if !yield(iterResult, err) {
				return
			}
			if cursor == "" {
				return
			}
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
				if body == nil {
					body = make(map[string]interface{})
				}
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
