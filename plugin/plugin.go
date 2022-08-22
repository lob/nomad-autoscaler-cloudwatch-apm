package plugin

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cloudwatchTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad-autoscaler/plugins"
	"github.com/hashicorp/nomad-autoscaler/plugins/apm"
	"github.com/hashicorp/nomad-autoscaler/plugins/base"
	"github.com/hashicorp/nomad-autoscaler/sdk"
)

const (
	// pluginName is the name of the plugin
	pluginName = "nomad-autoscaler-cloudwatch-apm"

	// configKeys represents the known configuration parameters required at
	// varying points throughout the plugins lifecycle.
	configKeyRegion       = "aws_region"
	configKeyAccessID     = "aws_access_key_id"
	configKeySecretKey    = "aws_secret_access_key"
	configKeySessionToken = "aws_session_token"

	// configValues are the default values used when a configuration key is not
	// supplied by the operator that are specific to the plugin.
	configValueRegionDefault = "us-east-1"
)

var (
	PluginID = plugins.PluginID{
		Name:       pluginName,
		PluginType: sdk.PluginTypeAPM,
	}

	PluginConfig = &plugins.InternalPluginConfig{
		Factory: func(l hclog.Logger) interface{} { return NewCloudWatchApmPlugin(l) },
	}

	pluginInfo = &base.PluginInfo{
		Name:       pluginName,
		PluginType: sdk.PluginTypeAPM,
	}
)

type CloudWatchApmPlugin struct {
	client    *cloudwatch.Client
	clientCtx context.Context
	config    map[string]string
	logger    hclog.Logger
}

func NewCloudWatchApmPlugin(log hclog.Logger) apm.APM {
	return &CloudWatchApmPlugin{
		logger: log,
	}
}

func (p *CloudWatchApmPlugin) SetConfig(config map[string]string) error {
	p.config = config
	p.clientCtx = context.Background()

	// Load our default AWS config. This handles pulling configuration from
	// default profiles and environment variables.
	cfg, err := awsconfig.LoadDefaultConfig(p.clientCtx)
	if err != nil {
		return fmt.Errorf("failed to load default AWS config: %v", err)
	}

	// If the operator has provided a configuration region, overwrite that set
	// by the AWS client.
	region, ok := config[configKeyRegion]
	if ok {
		p.logger.Debug("setting AWS region for client", "region", region)
		cfg.Region = region
	}

	// In the situation where the plugin is not running on an EC2 instance, nor
	// has the operator set an parameter, set the region to the default.
	if cfg.Region == "" {
		cfg.Region = configValueRegionDefault
	}

	// Attempt to pull access credentials for the AWS client from the user
	// supplied configuration. In order to use these static credentials both
	// the access key and secret key need to be present; the session token is
	// optional.
	// If not found, EC2RoleProvider will be instantiated instead.
	keyID := config[configKeyAccessID]
	secretKey := config[configKeySecretKey]
	session := config[configKeySessionToken]

	if keyID != "" && secretKey != "" {
		p.logger.Trace("setting AWS access credentials from config map")
		cfg.Credentials = credentials.NewStaticCredentialsProvider(keyID, secretKey, session)
	} else {
		p.logger.Trace("AWS access credentials empty - using EC2 instance role credentials instead")
		cfg.Credentials = aws.NewCredentialsCache(ec2rolecreds.New())
	}

	// Set up our AWS client.
	p.client = cloudwatch.NewFromConfig(cfg)

	return nil
}

func (p *CloudWatchApmPlugin) PluginInfo() (*base.PluginInfo, error) {
	return pluginInfo, nil
}

func (p *CloudWatchApmPlugin) Query(q string, r sdk.TimeRange) (sdk.TimestampedMetrics, error) {
	m, err := p.QueryMultiple(q, r)
	if err != nil {
		return nil, err
	}

	switch len(m) {
	case 0:
		return sdk.TimestampedMetrics{}, nil
	case 1:
		return m[0], nil
	default:
		return nil, fmt.Errorf("query returned %d metric streams, only 1 is expected", len(m))
	}
}

func (p *CloudWatchApmPlugin) QueryMultiple(q string, r sdk.TimeRange) ([]sdk.TimestampedMetrics, error) {
	ctx, cancel := context.WithTimeout(p.clientCtx, 10*time.Second)
	defer cancel()

	input := cloudwatch.GetMetricDataInput{
		StartTime: aws.Time(r.From),
		EndTime:   aws.Time(r.To),
		ScanBy:    cloudwatchTypes.ScanByTimestampDescending,
		MetricDataQueries: []cloudwatchTypes.MetricDataQuery{
			{
				Id:         aws.String("m1"),
				Expression: aws.String(q),
				ReturnData: aws.Bool(true),
				Period:     aws.Int32(1),
			},
		},
	}

	res, err := p.client.GetMetricData(ctx, &input)

	if err != nil {
		p.logger.Error(
			"Failed to get output",
			"error", err,
			"query", q,
			"from", r.From,
			"to", r.To,
		)
		return nil, fmt.Errorf("error querying metrics from cloudwtch: %v", err)
	}

	p.logger.Info(
		"Received Metric Data",
		"data", res,
		"query", q,
		"from", r.From,
		"to", r.To,
	)
	if len(res.MetricDataResults) == 0 {
		p.logger.Warn(
			"empty time series response from cloudwatch, try a wider query window",
			"query", q,
			"from", r.From,
			"to", r.To,
		)
		return nil, nil
	}

	var results []sdk.TimestampedMetrics
	for _, metric := range res.MetricDataResults {

		var result sdk.TimestampedMetrics

		for idx, value := range metric.Values {
			tm := sdk.TimestampedMetric{
				Timestamp: metric.Timestamps[idx],
				Value:     value,
			}
			result = append(result, tm)
		}

		results = append(results, result)
	}

	if len(results) == 0 {
		p.logger.Warn(
			"no data points found in time series response from cloudwatch, try a wider query window",
			"query", q,
			"from", r.From,
			"to", r.To,
		)
	}

	return results, nil
}
