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

type IterResponse struct {
	Response *http.Response
	Next     string
}

func (client *ApiClient) IterGet(
	path string,
	params *url.Values,
) iter.Seq2[*IterResponse, error] {

	type ScrollableResponse struct {
		Next string `json:"next"`
	}

	return func(yield func(*IterResponse, error) bool) {
		for {
			// Fire the request
			response, err := client.GetParams(
				path,
				params,
			)
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
			var scrollableResponse ScrollableResponse
			if err := json.Unmarshal(body, &scrollableResponse); err != nil {
				yield(
					nil,
					fmt.Errorf("failed to unmarshal response: %w", err),
				)
				return
			}

			// Replace the body and return the response
			response.Body = io.NopCloser(bytes.NewReader(body))
			if !yield(
				&IterResponse{
					Response: response,
					Next:     scrollableResponse.Next,
				},
				nil,
			) {
				return
			}

			// Was this the last page?
			if scrollableResponse.Next == "" {
				return
			}

			// Set the cursor for next page
			if params == nil {
				params = &url.Values{}
			}
			params.Set("from", scrollableResponse.Next)
		}
	}
}
