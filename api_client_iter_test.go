//go:build go1.23

package flareio

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIterGetParams(t *testing.T) {
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

	for response, err := range ct.apiClient.IterGet(
		"/leaksdb/sources",
		nil,
	) {
		lastPageIndex = lastPageIndex + 1
		if lastPageIndex > 5 {
			// We are going crazy here...
			break
		}
		assert.Nil(t, err, "iter yielded an error")
		assert.NotNil(t, response, "didn't get a response")
		nextTokens = append(nextTokens, response.Next)
	}

	assert.Equal(t, 3, lastPageIndex, "Didn't get the expected number of pages")
	assert.Equal(
		t,
		[]string{"second-page", "third-page", ""},
		nextTokens,
		"Didn't get the expected next tokens",
	)

}
