package main

import (
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad-autoscaler/plugins"
	cloudwatchPlugin "github.com/lob/nomad-autoscaler-cloudwatch-apm/plugin"
)

func main() {
	plugins.Serve(factory)
}

// factory returns a new instance of the Datadog APM plugin.
func factory(log hclog.Logger) interface{} {
	return cloudwatchPlugin.NewCloudWatchApmPlugin(log)
}
