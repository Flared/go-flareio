package main

import (
	"fmt"
	"io"
	"os"

	"github.com/Flared/go-flareio"
)

func main() {
	client := flareio.NewApiClient(
		os.Getenv("FLARE_API_KEY"),
	)
	resp, err := client.Get(
		"/tokens/test", nil,
	)
	if err != nil {
		fmt.Printf("failed to test token: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		fmt.Printf("failed to print response: %s\n", err)
		os.Exit(1)
	}
}
