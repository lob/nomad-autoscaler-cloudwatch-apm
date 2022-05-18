package plugin

import (
	"testing"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func TestApmPlugin_SetConfig(t *testing.T) {
	testCases := []struct {
		inputConfig          map[string]string
		expectOutput         error
		expectedContextKey   interface{}
		expectedContextValue interface{}
		name                 string
	}{
		// TODO: Add test cases
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cloudwatchPlugin := CloudWatchApmPlugin{logger: hclog.NewNullLogger()}

			// Perform the function call.
			actualOutput := cloudwatchPlugin.SetConfig(tc.inputConfig)
			assert.Equal(t, tc.expectOutput, actualOutput, tc.name)

			// Check the stored context and the client. If we expect to have a
			// non-nil context then we should have a non-nil client and vice
			// versa.
			if tc.expectedContextValue != nil {
				// assert.Equal(t, tc.expectedContextValue, cloudwatchPlugin.clientCtx.Value(tc.expectedContextKey), tc.name)
				assert.NotNil(t, cloudwatchPlugin.client, tc.name)
			} else {
				// assert.Nil(t, cloudwatchPlugin.clientCtx, tc.name)
				assert.Nil(t, cloudwatchPlugin.client, tc.name)
			}
		})
	}
}
