package cli

import (
	"strings"
	"testing"
)

func TestAgentsHelp_Embedded(t *testing.T) {
	if agentsHelp == "" {
		t.Fatal("agentsHelp is empty - embed failed")
	}
}

func TestSchemasHelp_Embedded(t *testing.T) {
	if schemasHelp == "" {
		t.Fatal("schemasHelp is empty - embed failed")
	}
}

func TestAgentsHelp_ContainsExpectedSections(t *testing.T) {
	sections := []string{
		"# ajira Agent Reference",
		"## Key Behaviours",
		"## Find Issues",
		"## View Issue",
		"## Create Issue",
		"## Modify Issue",
		"## Comments",
		"## Chaining",
	}

	for _, section := range sections {
		if !strings.Contains(agentsHelp, section) {
			t.Errorf("agentsHelp missing section: %s", section)
		}
	}
}

func TestSchemasHelp_ContainsExpectedSections(t *testing.T) {
	sections := []string{
		"# ajira JSON Schemas",
		"## me",
		"## project list",
		"## issue list",
		"## issue view",
		"## issue create",
	}

	for _, section := range sections {
		if !strings.Contains(schemasHelp, section) {
			t.Errorf("schemasHelp missing section: %s", section)
		}
	}
}

func TestAgentsHelp_TokenEfficient(t *testing.T) {
	// Rough token estimate: words * 1.3
	words := len(strings.Fields(agentsHelp))
	estimatedTokens := int(float64(words) * 1.3)

	if estimatedTokens > 2000 {
		t.Errorf("agentsHelp exceeds 2000 token target: ~%d tokens (%d words)", estimatedTokens, words)
	}
}

func TestAgentsHelp_ReferencesSchemas(t *testing.T) {
	if !strings.Contains(agentsHelp, "ajira help schemas") {
		t.Error("agentsHelp should reference 'ajira help schemas'")
	}
}

func TestHelpCommands_Registered(t *testing.T) {
	// Check that help command has agents and schemas subcommands
	var hasAgents, hasSchemas bool

	for _, cmd := range helpCmd.Commands() {
		switch cmd.Name() {
		case "agents":
			hasAgents = true
		case "schemas":
			hasSchemas = true
		}
	}

	if !hasAgents {
		t.Error("help command missing 'agents' subcommand")
	}
	if !hasSchemas {
		t.Error("help command missing 'schemas' subcommand")
	}
}

func TestHelpAgentsCmd_Properties(t *testing.T) {
	if helpAgentsCmd.Use != "agents" {
		t.Errorf("expected Use 'agents', got %s", helpAgentsCmd.Use)
	}
	if helpAgentsCmd.Short == "" {
		t.Error("helpAgentsCmd.Short is empty")
	}
}

func TestHelpSchemasCmd_Properties(t *testing.T) {
	if helpSchemasCmd.Use != "schemas" {
		t.Errorf("expected Use 'schemas', got %s", helpSchemasCmd.Use)
	}
	if helpSchemasCmd.Short == "" {
		t.Error("helpSchemasCmd.Short is empty")
	}
}
