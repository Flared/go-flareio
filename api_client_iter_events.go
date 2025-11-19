//go:build go1.23

package flareio

import (
	"iter"
	"net/http"
	"net/url"
)

// IterEvent contains results for a given page.
type IterEventsResult struct {
	// Response associated with the fetched page.
	//
	// The response's body must be closed.
	Response *http.Response

	// Next is the token to be used to fetch the next page.
	Next string
}

func (client *ApiClient) IterEventsPostJson(
	pagesPath string,
	eventsPath string,
	parms *url.Values,
	body map[string]interface{},
) iter.Seq2[*IterEventsResult, error] {
	return func(yield func(*IterEventsResult, error) bool) {
	}
}
