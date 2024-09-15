//go:build go1.23

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Flared/go-flareio"
)

func main() {
	client := flareio.NewApiClient(
		os.Getenv("FLARE_API_KEY"),
	)

	fetchedPages := 0

	for result, err := range client.IterGet(
		"/leaksdb/v2/sources", nil,
	) {
		func(result *flareio.IterResult, err error) {
			// Rate Limiting
			time.Sleep(time.Second * 1)

			if err != nil {
				fmt.Printf("unexpected error: %s\n", err)
				os.Exit(1)
			}

			// Handle the response...
			defer result.Response.Body.Close()

			// Print the status
			fetchedPages = fetchedPages + 1
			fmt.Printf(
				"Fetched %d page(s) of LeaksDB Sources, next=%s\n",
				fetchedPages,
				result.Next,
			)
		}(result, err)
	}

}
