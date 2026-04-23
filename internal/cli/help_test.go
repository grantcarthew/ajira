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

func TestAgileHelp_Embedded(t *testing.T) {
	if agileHelp == "" {
		t.Fatal("agileHelp is empty - embed failed")
	}
}

func TestMarkdownHelp_Embedded(t *testing.T) {
	if markdownHelp == "" {
		t.Fatal("markdownHelp is empty - embed failed")
	}
}

func TestMarkdownHelp_ContainsExpectedSections(t *testing.T) {
	sections := []string{
		"# Markdown Formatting Reference",
		"## Cheatsheet",
		"## Tables",
		"## Code Blocks",
		"## Gotchas",
		"## Example",
	}

	for _, section := range sections {
		if !strings.Contains(markdownHelp, section) {
			t.Errorf("markdownHelp missing section: %s", section)
		}
	}
}

func TestMarkdownHelp_TokenEfficient(t *testing.T) {
	words := len(strings.Fields(markdownHelp))
	estimatedTokens := int(float64(words) * 1.3)

	if estimatedTokens > 800 {
		t.Errorf("markdownHelp exceeds 800 token target: ~%d tokens (%d words)", estimatedTokens, words)
	}
}

func TestAgileHelp_ContainsExpectedSections(t *testing.T) {
	sections := []string{
		"# ajira Agile",
		"## Commands",
		"ajira board list",
		"ajira sprint list",
		"ajira sprint add",
		"ajira epic list",
		"ajira epic create",
		"ajira epic add",
		"ajira epic remove",
	}

	for _, section := range sections {
		if !strings.Contains(agileHelp, section) {
			t.Errorf("agileHelp missing section: %s", section)
		}
	}
}

func TestAgileHelp_TokenEfficient(t *testing.T) {
	words := len(strings.Fields(agileHelp))
	estimatedTokens := int(float64(words) * 1.3)

	if estimatedTokens > 500 {
		t.Errorf("agileHelp exceeds 500 token target: ~%d tokens (%d words)", estimatedTokens, words)
	}
}

func TestAgentsHelp_ContainsExpectedSections(t *testing.T) {
	sections := []string{
		"# ajira Jira CLI",
		"## Markdown",
		"## Core Commands",
		"## Other Commands",
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
		"me:",
		"project list:",
		"issue list:",
		"issue view:",
		"issue create:",
		"issue comment list:",
		"issue link list:",
		"issue link types:",
		"issue link add:",
		"issue attachment list:",
		"epic list:",
		"sprint list:",
	}

	for _, section := range sections {
		if !strings.Contains(schemasHelp, section) {
			t.Errorf("schemasHelp missing section: %s", section)
		}
	}
}

func TestSchemasHelp_TokenEfficient(t *testing.T) {
	words := len(strings.Fields(schemasHelp))
	estimatedTokens := int(float64(words) * 1.3)

	if estimatedTokens > 800 {
		t.Errorf("schemasHelp exceeds 800 token target: ~%d tokens (%d words)", estimatedTokens, words)
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

func TestAgentsHelp_ReferencesTopics(t *testing.T) {
	for _, topic := range []string{"schemas", "markdown", "agile"} {
		if !strings.Contains(agentsHelp, topic) {
			t.Errorf("agentsHelp should reference '%s'", topic)
		}
	}
}

func TestHelpCommands_Registered(t *testing.T) {
	var hasAgents, hasSchemas, hasMarkdown, hasAgile bool

	for _, cmd := range helpCmd.Commands() {
		switch cmd.Name() {
		case "agents":
			hasAgents = true
		case "schemas":
			hasSchemas = true
		case "markdown":
			hasMarkdown = true
		case "agile":
			hasAgile = true
		}
	}

	if !hasAgents {
		t.Error("help command missing 'agents' subcommand")
	}
	if !hasSchemas {
		t.Error("help command missing 'schemas' subcommand")
	}
	if !hasMarkdown {
		t.Error("help command missing 'markdown' subcommand")
	}
	if !hasAgile {
		t.Error("help command missing 'agile' subcommand")
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
