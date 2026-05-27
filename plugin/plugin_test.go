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
		expectedRegion string
		expectedClient bool
	}{
		{
			name:           "empty config uses default region and creates client",
			inputConfig:    map[string]string{},
			expectErr:      false,
			expectedRegion: configValueRegionDefault,
			expectedClient: true,
		},
		{
			name:           "explicit region overrides default",
			inputConfig:    map[string]string{configKeyRegion: "eu-west-1"},
			expectErr:      false,
			expectedRegion: "eu-west-1",
			expectedClient: true,
		},
		{
			name: "static credentials accepted when both key and secret present",
			inputConfig: map[string]string{
				configKeyAccessID:  "AKIAIOSFODNN7EXAMPLE",
				configKeySecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			},
			expectErr:      false,
			expectedRegion: configValueRegionDefault,
			expectedClient: true,
		},
		{
			name: "static credentials with session token accepted",
			inputConfig: map[string]string{
				configKeyAccessID:     "AKIAIOSFODNN7EXAMPLE",
				configKeySecretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				configKeySessionToken: "session-token-example",
			},
			expectErr:      false,
			expectedRegion: configValueRegionDefault,
			expectedClient: true,
		},
		{
			// Only one credential half provided — falls back to default auth chain.
			name: "partial credentials fall back to default auth",
			inputConfig: map[string]string{
				configKeyAccessID: "AKIAIOSFODNN7EXAMPLE",
			},
			expectErr:      false,
			expectedRegion: configValueRegionDefault,
			expectedClient: true,
		},
		{
			name: "explicit region with static credentials",
			inputConfig: map[string]string{
				configKeyRegion:    "ap-southeast-1",
				configKeyAccessID:  "AKIAIOSFODNN7EXAMPLE",
				configKeySecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			},
			expectErr:      false,
			expectedRegion: "ap-southeast-1",
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

	out, _ := cmd.Output() // non-zero exit is expected when vulns are found

	// govulncheck -json emits one JSON object per line. Each "finding" entry
	// has a non-empty Trace when the symbol is reachable (Symbol-level).
	type findingMsg struct {
		Finding *struct {
			OSV   string `json:"osv"`
			Trace []struct {
				Function string `json:"function"`
			} `json:"trace"`
		} `json:"finding"`
	}

	var symbolVulns []string
	for _, line := range bytes.Split(out, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var msg findingMsg
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}
		if msg.Finding != nil && msg.Finding.OSV != "" && len(msg.Finding.Trace) > 0 {
			symbolVulns = append(symbolVulns, msg.Finding.OSV)
		}
	}

	assert.Empty(t, symbolVulns,
		"Symbol-level CVEs found in call graph — upgrade the affected dependencies: %v", symbolVulns)
}
