package plugin

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApmPlugin_SetConfig(t *testing.T) {
	testCases := []struct {
		name           string
		inputConfig    map[string]string
		expectErr      bool
		expectedClient bool
	}{
		{
			name:           "empty config uses default region and creates client",
			inputConfig:    map[string]string{},
			expectedClient: true,
		},
		{
			name:           "explicit region overrides default",
			inputConfig:    map[string]string{configKeyRegion: "eu-west-1"},
			expectedClient: true,
		},
		{
			name: "static credentials accepted when both key and secret present",
			inputConfig: map[string]string{
				configKeyAccessID:  "AKIAIOSFODNN7EXAMPLE",
				configKeySecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			},
			expectedClient: true,
		},
		{
			name: "static credentials with session token accepted",
			inputConfig: map[string]string{
				configKeyAccessID:     "AKIAIOSFODNN7EXAMPLE",
				configKeySecretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				configKeySessionToken: "session-token-example",
			},
			expectedClient: true,
		},
		{
			// Only one credential half provided — falls back to default auth chain.
			name: "partial credentials fall back to default auth",
			inputConfig: map[string]string{
				configKeyAccessID: "AKIAIOSFODNN7EXAMPLE",
			},
			expectedClient: true,
		},
		{
			name: "explicit region with static credentials",
			inputConfig: map[string]string{
				configKeyRegion:    "ap-southeast-1",
				configKeyAccessID:  "AKIAIOSFODNN7EXAMPLE",
				configKeySecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			},
			expectedClient: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &CloudWatchApmPlugin{logger: hclog.NewNullLogger()}

			err := p.SetConfig(tc.inputConfig)

			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, p.clientCtx, "clientCtx should be set")

			if tc.expectedClient {
				assert.NotNil(t, p.client, "CloudWatch client should be initialised")
			} else {
				assert.Nil(t, p.client)
			}

			// Verify the stored config map is exactly what was passed in.
			assert.Equal(t, tc.inputConfig, p.config)
		})
	}
}

func TestApmPlugin_PluginInfo(t *testing.T) {
	p := &CloudWatchApmPlugin{logger: hclog.NewNullLogger()}
	info, err := p.PluginInfo()
	require.NoError(t, err)
	assert.Equal(t, pluginName, info.Name)
}

// TestNoSymbolVulnerabilities shells out to govulncheck and fails if any
// Symbol-level CVEs are reachable from the binary's call graph.
//
// Symbol-level means the vulnerable function is in the actual call graph, not
// merely present in a transitive dependency. This is the tier that represents
// real runtime exposure. The test is skipped if govulncheck is unavailable
// (e.g., local env without Go tools), but will run in CI via `go tool`.
func TestNoSymbolVulnerabilities(t *testing.T) {
	// Resolve module root (one level up from the plugin package).
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")
	moduleRoot := filepath.Join(filepath.Dir(thisFile), "..")

	// Prefer `go tool govulncheck` (declared in go.mod tool directive).
	// Fall back to PATH for local environments where `go tool` may not cache it.
	govulncheck, err := exec.LookPath("govulncheck")
	if err != nil {
		goExe, err2 := exec.LookPath("go")
		if err2 != nil {
			t.Skip("govulncheck not in PATH and go not found; skipping")
		}
		govulncheck = goExe
	}

	var cmd *exec.Cmd
	if filepath.Base(govulncheck) == "go" {
		cmd = exec.Command(govulncheck, "tool", "govulncheck", "-json", "./...")
	} else {
		cmd = exec.Command(govulncheck, "-json", "./...")
	}
	cmd.Dir = moduleRoot

	out, runErr := cmd.Output()

	// If the command produced no output it likely failed for a non-CVE reason
	// (network error fetching the vuln DB, permission issue, etc.). Skip rather
	// than silently passing — the caller should investigate.
	if len(out) == 0 {
		stderr := ""
		if ee, ok2 := runErr.(*exec.ExitError); ok2 {
			stderr = string(ee.Stderr)
		}
		t.Skipf("govulncheck produced no output (non-CVE failure); stderr: %s", stderr)
	}

	// govulncheck -json emits a stream of pretty-printed JSON objects. Each
	// "finding" object has a non-empty Trace when the symbol is reachable
	// (Symbol-level). Use a streaming decoder so multi-line objects are handled
	// correctly.
	type findingMsg struct {
		Finding *struct {
			OSV   string `json:"osv"`
			Trace []struct {
				Function string `json:"function"`
			} `json:"trace"`
		} `json:"finding"`
	}

	var symbolVulns []string
	decoder := json.NewDecoder(bytes.NewReader(out))
	for decoder.More() {
		var msg findingMsg
		if err := decoder.Decode(&msg); err != nil {
			// Warn rather than silently skip — a decode failure may mean the
			// govulncheck JSON schema changed, which would cause false negatives.
			t.Logf("govulncheck: unexpected decode failure (schema change?): %v", err)
			continue
		}
		if msg.Finding != nil && msg.Finding.OSV != "" && len(msg.Finding.Trace) > 0 {
			symbolVulns = append(symbolVulns, msg.Finding.OSV)
		}
	}

	assert.Empty(t, symbolVulns,
		"Symbol-level CVEs found in call graph — upgrade the affected dependencies: %v", symbolVulns)
}
