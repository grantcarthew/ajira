package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// BatchResult represents the outcome of a single batch operation.
type BatchResult struct {
	Key     string `json:"key"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// BatchSummary represents the overall batch operation outcome.
type BatchSummary struct {
	Results   []BatchResult `json:"results"`
	Total     int           `json:"total"`
	Succeeded int           `json:"succeeded"`
	Failed    int           `json:"failed"`
}

// ReadKeysFromStdin reads issue keys from stdin (one per line).
func ReadKeysFromStdin() ([]string, error) {
	var keys []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		key := strings.TrimSpace(scanner.Text())
		if key != "" {
			keys = append(keys, key)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading stdin: %w", err)
	}
	return keys, nil
}

// PrintBatchResults prints batch operation results.
// Returns an ExitError with ExitPartial if there were any failures.
func PrintBatchResults(results []BatchResult) error {
	summary := BatchSummary{
		Results: results,
		Total:   len(results),
	}

	for _, r := range results {
		if r.Success {
			summary.Succeeded++
		} else {
			summary.Failed++
		}
	}

	if JSONOutput() {
		output, _ := json.MarshalIndent(summary, "", "  ")
		fmt.Println(string(output))
	} else {
		// Print individual results
		for _, r := range results {
			if r.Success {
				fmt.Printf("%s: success\n", r.Key)
			} else {
				fmt.Printf("%s: failed - %s\n", r.Key, r.Error)
			}
		}
		// Print summary
		fmt.Printf("\n%d processed: %d succeeded, %d failed\n", summary.Total, summary.Succeeded, summary.Failed)
	}

	if summary.Failed > 0 {
		if summary.Succeeded > 0 {
			return NewExitError(ExitPartial, fmt.Errorf("partial failure: %d of %d failed", summary.Failed, summary.Total))
		}
		return NewExitError(ExitAPIError, fmt.Errorf("all %d operations failed", summary.Total))
	}

	return nil
}

// PrintDryRunBatch prints what would happen for a batch operation.
func PrintDryRunBatch(keys []string, action string) {
	if JSONOutput() {
		type dryRunItem struct {
			Key    string `json:"key"`
			Action string `json:"action"`
		}
		items := make([]dryRunItem, len(keys))
		for i, key := range keys {
			items[i] = dryRunItem{Key: key, Action: action}
		}
		output, _ := json.MarshalIndent(items, "", "  ")
		fmt.Println(string(output))
	} else {
		for _, key := range keys {
			fmt.Printf("Would %s %s\n", action, key)
		}
	}
}

// PrintDryRun prints what would happen for a single operation.
func PrintDryRun(action string) {
	if JSONOutput() {
		result := map[string]string{"action": action}
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Printf("Would %s\n", action)
	}
}

// PrintSuccess prints a success message unless quiet mode is enabled.
func PrintSuccess(message string) {
	if !Quiet() {
		fmt.Println(message)
	}
}

// PrintSuccessJSON prints JSON output unless quiet mode is enabled.
func PrintSuccessJSON(v any) {
	if !Quiet() {
		output, _ := json.MarshalIndent(v, "", "  ")
		fmt.Println(string(output))
	}
}
