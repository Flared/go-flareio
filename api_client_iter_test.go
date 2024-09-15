//go:build go1.23

package flareio

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIterGet(t *testing.T) {
	ct := newClientTest(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/leaksdb/sources", r.URL.Path)

			cursor := r.URL.Query().Get("from")
			if cursor == "" {
				w.Write([]byte(`{"next":"second-page", "items": []}`))
			} else if cursor == "second-page" {
				w.Write([]byte(`{"next":"third-page", "items": []}`))
			} else {
				w.Write([]byte(`{"next": null, "items": []}`))
			}
		}),
	)
	defer ct.Close()

	lastPageIndex := 0
	nextTokens := []string{}

	for result, err := range ct.apiClient.IterGet(
		"/leaksdb/sources",
		nil,
	) {
		lastPageIndex = lastPageIndex + 1
		if lastPageIndex > 5 {
			// We are going crazy here...
			break
		}
		assert.Nil(t, err, "iter yielded an error")
		assert.NotNil(t, result, "didn't get a result")
		nextTokens = append(nextTokens, result.Next)
	}

	assert.Equal(t, 3, lastPageIndex, "Didn't get the expected number of pages")
	assert.Equal(
		t,
		[]string{"second-page", "third-page", ""},
		nextTokens,
		"Didn't get the expected next tokens",
	)

}

func TestIterGetBadResponse(t *testing.T) {
	ct := newClientTest(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/leaksdb/sources", r.URL.Path)
			w.Write([]byte(`{"next": 11, "items": []}`))
		}),
	)
	defer ct.Close()

	lastPageIndex := 0

	for result, err := range ct.apiClient.IterGet(
		"/leaksdb/sources",
		nil,
	) {
		lastPageIndex = lastPageIndex + 1
		if lastPageIndex > 2 {
			// We are going crazy here...
			break
		}
		assert.ErrorContains(t, err, "failed to unmarshal", "Bad next token should trigger an error")
		assert.Nil(t, result, "bad response should not contain a result")
	}

	assert.Equal(t, 1, lastPageIndex, "Didn't get the expected number of pages")
}

func TestIterPostJson(t *testing.T) {
	ct := newClientTest(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/leaksdb/sources", r.URL.Path)
			assert.Equal(t, "value1", r.URL.Query().Get("param1"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			type PagedRequest struct {
				From string `json:"from"`
			}
			var pagedRequest PagedRequest
			if err := json.NewDecoder(r.Body).Decode(&pagedRequest); !assert.NoError(t, err, "Error decoding posted JSON") {
				return
			}

			cursor := pagedRequest.From
			if cursor == "" {
				w.Write([]byte(`{"next":"second-page", "items": []}`))
			} else if cursor == "second-page" {
				w.Write([]byte(`{"next":"third-page", "items": []}`))
			} else {
				w.Write([]byte(`{"next": null, "items": []}`))
			}
		}),
	)
	defer ct.Close()

	lastPageIndex := 0
	nextTokens := []string{}

	for result, err := range ct.apiClient.IterPostJson(
		"/leaksdb/sources",
		&url.Values{
			"param1": []string{"value1"},
		},
		map[string]interface{}{
			"some_param": "hello",
		},
	) {
		lastPageIndex = lastPageIndex + 1
		if lastPageIndex > 5 {
			// We are going crazy here...
			break
		}
		assert.Nil(t, err, "iter yielded an error")
		assert.NotNil(t, result, "didn't get a result")
		nextTokens = append(nextTokens, result.Next)
	}

	assert.Equal(t, 3, lastPageIndex, "Didn't get the expected number of pages")
	assert.Equal(
		t,
		[]string{"second-page", "third-page", ""},
		nextTokens,
		"Didn't get the expected next tokens",
	)

}

func TestIterPostJsonNilMap(t *testing.T) {
	requestsReceived := 0

	ct := newClientTest(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestsReceived = requestsReceived + 1
			if requestsReceived == 1 {
				w.Write([]byte(`{"next": "second-page"}`))
			} else {
				w.Write([]byte(`{"next": null}`))
			}
		}),
	)
	defer ct.Close()

	requestsSent := 0
	nextTokens := []string{}

	for result, err := range ct.apiClient.IterPostJson(
		"/leaksdb/sources",
		nil,
		nil,
	) {
		requestsSent = requestsSent + 1
		if requestsSent > 2 {
			// We are going crazy here...
			break
		}
		assert.Nil(t, err, "iter yielded an error")
		assert.NotNil(t, result, "didn't get a result")
		nextTokens = append(nextTokens, result.Next)
	}

	assert.Equal(t, 2, requestsSent, "Didn't get the expected number of pages")
}
