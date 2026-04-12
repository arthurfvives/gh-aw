//go:build !integration

package workflow

import (
	"strings"
	"testing"

	"github.com/github/gh-aw/pkg/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildPullAWFContainersStep verifies that the detection job's container
// download step includes the correct images, matching what the execution step needs.
func TestBuildPullAWFContainersStep(t *testing.T) {
	awfImageTag := strings.TrimPrefix(string(constants.DefaultFirewallVersion), "v")

	t.Run("includes cli-proxy image when features has cli-proxy enabled", func(t *testing.T) {
		compiler := &Compiler{actionMode: ActionModeRelease}
		data := &WorkflowData{
			AI:       "claude",
			Features: map[string]any{"cli-proxy": true},
			SandboxConfig: &SandboxConfig{
				Agent: &AgentSandboxConfig{
					Type: SandboxTypeAWF,
				},
			},
		}

		steps := compiler.buildPullAWFContainersStep(data)
		require.NotNil(t, steps, "expected non-nil steps for detection container download")

		joined := strings.Join(steps, "")
		cliProxyImage := constants.DefaultFirewallRegistry + "/cli-proxy:" + awfImageTag
		assert.Contains(t, joined, "cli-proxy",
			"detection download step should include cli-proxy image when feature is enabled, got: %s", joined)
		assert.Contains(t, joined, cliProxyImage,
			"detection download step should include cli-proxy image %q", cliProxyImage)
	})

	t.Run("excludes cli-proxy image when features does not have cli-proxy", func(t *testing.T) {
		compiler := &Compiler{actionMode: ActionModeRelease}
		data := &WorkflowData{
			AI: "claude",
			SandboxConfig: &SandboxConfig{
				Agent: &AgentSandboxConfig{
					Type: SandboxTypeAWF,
				},
			},
		}

		steps := compiler.buildPullAWFContainersStep(data)
		require.NotNil(t, steps, "expected non-nil steps for detection container download")

		joined := strings.Join(steps, "")
		assert.NotContains(t, joined, "cli-proxy",
			"detection download step should NOT include cli-proxy image when feature is absent")
	})

	t.Run("always includes squid agent and api-proxy images", func(t *testing.T) {
		compiler := &Compiler{actionMode: ActionModeRelease}
		data := &WorkflowData{
			AI: "claude",
			SandboxConfig: &SandboxConfig{
				Agent: &AgentSandboxConfig{
					Type: SandboxTypeAWF,
				},
			},
		}

		steps := compiler.buildPullAWFContainersStep(data)
		require.NotNil(t, steps, "expected non-nil steps for detection container download")

		joined := strings.Join(steps, "")
		assert.Contains(t, joined, "/squid:"+awfImageTag,
			"detection download step should include squid image")
		assert.Contains(t, joined, "/agent:"+awfImageTag,
			"detection download step should include agent image")
		assert.Contains(t, joined, "/api-proxy:"+awfImageTag,
			"detection download step should include api-proxy image")
	})

	t.Run("propagates ActionCache for digest pinning", func(t *testing.T) {
		compiler := &Compiler{actionMode: ActionModeRelease}
		squidImage := constants.DefaultFirewallRegistry + "/squid:" + awfImageTag
		data := &WorkflowData{
			AI: "claude",
			SandboxConfig: &SandboxConfig{
				Agent: &AgentSandboxConfig{
					Type: SandboxTypeAWF,
				},
			},
			ActionCache: &ActionCache{
				ContainerPins: map[string]ContainerPin{
					squidImage: {
						PinnedImage: squidImage + "@sha256:abc123",
					},
				},
			},
		}

		steps := compiler.buildPullAWFContainersStep(data)
		require.NotNil(t, steps, "expected non-nil steps for detection container download")

		joined := strings.Join(steps, "")
		assert.Contains(t, joined, "@sha256:abc123",
			"detection download step should include digest-pinned image reference")
	})
}
