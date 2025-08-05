package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	baseURL := "http://localhost:8080/api/v1/servers/export"

	resp, err := http.Get(baseURL)
	if err != nil {
		fmt.Printf("failed to call export API: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("export API failed (Status: %d): %s\n", resp.StatusCode, string(body))
		return
	}

	fmt.Printf("📁 Export completed - file saved to server exports directory\n")
}
