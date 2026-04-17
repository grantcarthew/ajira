package cli

import (
	"io"
	"os"
)

// readText reads text content from file, stdin, or falls back to body.
// If file is "-", reads from stdin. If file is a path, reads that file.
// If file is empty, returns body.
func readText(file, body string) (string, error) {
	if file == "" {
		return body, nil
	}
	if file == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
