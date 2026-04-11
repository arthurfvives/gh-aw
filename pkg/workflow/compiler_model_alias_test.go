//go:build !integration

package workflow

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompilerResolveModelAlias_NoRepoConfig(t *testing.T) {
	compiler := NewCompiler()
	resolved, isAlias := compiler.resolveModelAlias("copilot", "small")
	assert.Equal(t, "small", resolved, "should return original model when no repo config")
	assert.False(t, isAlias, "should not be an alias when no repo config")
}

func TestCompilerResolveModelAlias_WithRepoConfig(t *testing.T) {
	compiler := NewCompiler()
	compiler.SetRepoConfig(&RepoConfig{
		Models: map[string]map[string]string{
			"copilot": {
				"small": "gpt-5.4-codex-mini",
				"large": "gpt-5",
			},
			"claude": {
				"fast": "claude-haiku-4.5",
			},
		},
	})

	tests := []struct {
		name         string
		engineID     string
		model        string
		wantResolved string
		wantIsAlias  bool
	}{
		{
			name:         "copilot small resolves",
			engineID:     "copilot",
			model:        "small",
			wantResolved: "gpt-5.4-codex-mini",
			wantIsAlias:  true,
		},
		{
			name:         "copilot large resolves",
			engineID:     "copilot",
			model:        "large",
			wantResolved: "gpt-5",
			wantIsAlias:  true,
		},
		{
			name:         "claude fast resolves",
			engineID:     "claude",
			model:        "fast",
			wantResolved: "claude-haiku-4.5",
			wantIsAlias:  true,
		},
		{
			name:         "unknown alias passes through",
			engineID:     "copilot",
			model:        "unknown",
			wantResolved: "unknown",
			wantIsAlias:  false,
		},
		{
			name:         "explicit model passes through",
			engineID:     "copilot",
			model:        "gpt-5.4-codex-mini",
			wantResolved: "gpt-5.4-codex-mini",
			wantIsAlias:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, isAlias := compiler.resolveModelAlias(tt.engineID, tt.model)
			assert.Equal(t, tt.wantResolved, resolved, "resolved model should match")
			assert.Equal(t, tt.wantIsAlias, isAlias, "isAlias flag should match")
		})
	}
}

func TestGenerateCreateAwInfo_ModelAlias(t *testing.T) {
	compiler := NewCompiler()
	compiler.SetRepoConfig(&RepoConfig{
		Models: map[string]map[string]string{
			"copilot": {
				"small": "gpt-5.4-codex-mini",
			},
		},
	})

	data := &WorkflowData{
		Name: "test-workflow",
		EngineConfig: &EngineConfig{
			ID:    "copilot",
			Model: "small",
		},
		Tools: make(map[string]any),
	}

	engine := &CopilotEngine{}

	var yaml strings.Builder
	compiler.generateCreateAwInfo(&yaml, data, engine)
	output := yaml.String()

	// Should contain the resolved model, not the alias
	assert.Contains(t, output, `GH_AW_INFO_MODEL: "gpt-5.4-codex-mini"`, "should emit resolved model")
	assert.Contains(t, output, `GH_AW_INFO_MODEL_ALIAS: "small"`, "should emit model alias")
}

func TestGenerateCreateAwInfo_NoAlias(t *testing.T) {
	compiler := NewCompiler()
	compiler.SetRepoConfig(&RepoConfig{
		Models: map[string]map[string]string{
			"copilot": {
				"small": "gpt-5.4-codex-mini",
			},
		},
	})

	data := &WorkflowData{
		Name: "test-workflow",
		EngineConfig: &EngineConfig{
			ID:    "copilot",
			Model: "gpt-5",
		},
		Tools: make(map[string]any),
	}

	engine := &CopilotEngine{}

	var yaml strings.Builder
	compiler.generateCreateAwInfo(&yaml, data, engine)
	output := yaml.String()

	// Should contain the explicit model and no alias
	assert.Contains(t, output, `GH_AW_INFO_MODEL: "gpt-5"`, "should emit explicit model")
	assert.NotContains(t, output, "GH_AW_INFO_MODEL_ALIAS", "should not emit model alias when not an alias")
}

func TestGenerateCreateAwInfo_NoRepoConfig(t *testing.T) {
	compiler := NewCompiler()
	// No repo config set

	data := &WorkflowData{
		Name: "test-workflow",
		EngineConfig: &EngineConfig{
			ID:    "copilot",
			Model: "small",
		},
		Tools: make(map[string]any),
	}

	engine := &CopilotEngine{}

	var yaml strings.Builder
	compiler.generateCreateAwInfo(&yaml, data, engine)
	output := yaml.String()

	// Should pass through "small" as-is since no repo config
	assert.Contains(t, output, `GH_AW_INFO_MODEL: "small"`, "should emit model as-is without repo config")
	assert.NotContains(t, output, "GH_AW_INFO_MODEL_ALIAS", "should not emit model alias without repo config")
}

func TestGenerateWorkflowHeader_ModelAliasResolution(t *testing.T) {
	compiler := NewCompiler()
	compiler.SetRepoConfig(&RepoConfig{
		Models: map[string]map[string]string{
			"copilot": {
				"small": "gpt-5.4-codex-mini",
			},
		},
	})

	data := &WorkflowData{
		Name: "test-workflow",
		EngineConfig: &EngineConfig{
			ID:    "copilot",
			Model: "small",
		},
		Tools: make(map[string]any),
	}

	var yaml strings.Builder
	compiler.generateWorkflowHeader(&yaml, data, "abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abcd", nil, nil)
	output := yaml.String()

	// The lock metadata should contain the resolved model, not the alias
	require.Contains(t, output, "gh-aw-metadata", "should contain metadata line")
	assert.Contains(t, output, "gpt-5.4-codex-mini", "lock metadata should contain resolved model")
	assert.NotContains(t, output, `"agent_model":"small"`, "lock metadata should not contain alias as model")
}
