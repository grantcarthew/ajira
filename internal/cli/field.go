package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// FieldInfo represents a Jira field.
type FieldInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Custom bool   `json:"custom"`
	Type   string `json:"type,omitempty"`
}

// fieldResponse matches the Jira field API response.
type fieldResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Custom bool   `json:"custom"`
	Schema *struct {
		Type string `json:"type"`
	} `json:"schema"`
}

var fieldCustomOnly bool

var fieldCmd = &cobra.Command{
	Use:   "field",
	Short: "Manage Jira fields",
	Long:  "Commands for listing and discovering Jira fields.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var fieldListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available fields",
	Long: `List all available Jira fields including system and custom fields.

Field IDs (e.g., customfield_10001) can be used in JQL queries or API calls.`,
	Example: `  ajira field list              # List all fields
  ajira field list --custom     # List only custom fields
  ajira field list --json       # Output as JSON`,
	SilenceUsage: true,
	RunE:         runFieldList,
}

func init() {
	fieldListCmd.Flags().BoolVar(&fieldCustomOnly, "custom", false, "Show only custom fields")

	fieldCmd.AddCommand(fieldListCmd)
	rootCmd.AddCommand(fieldCmd)
}

func runFieldList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	fields, err := fetchFields(ctx, client)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to fetch fields: %v", err)
	}

	// Filter to custom fields only if requested
	if fieldCustomOnly {
		var customFields []FieldInfo
		for _, f := range fields {
			if f.Custom {
				customFields = append(customFields, f)
			}
		}
		fields = customFields
	}

	// Sort by name
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})

	if JSONOutput() {
		output, err := json.MarshalIndent(fields, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		if len(fields) == 0 {
			fmt.Println("No fields found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tTYPE\tCUSTOM")
		for _, f := range fields {
			custom := ""
			if f.Custom {
				custom = "Yes"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", f.ID, f.Name, f.Type, custom)
		}
		w.Flush()
	}

	return nil
}

func fetchFields(ctx context.Context, client *api.Client) ([]FieldInfo, error) {
	body, err := client.Get(ctx, "/field")
	if err != nil {
		return nil, err
	}

	var fields []fieldResponse
	if err := json.Unmarshal(body, &fields); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result := make([]FieldInfo, len(fields))
	for i, f := range fields {
		fieldType := ""
		if f.Schema != nil {
			fieldType = f.Schema.Type
		}
		result[i] = FieldInfo{
			ID:     f.ID,
			Name:   f.Name,
			Custom: f.Custom,
			Type:   fieldType,
		}
	}

	return result, nil
}
