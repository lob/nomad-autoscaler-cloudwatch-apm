# Nomad Autoscaler CloudWatch APM

A Nomad Autoscaler APM plugin to scale Nomad jobs using CloudWatch Metrics.

**This software is in early development and should be used with caution.**

*Feedback and contributions are welcome :-)*

## Plugin Configuration

To use the plugin you will need to download the binary to the client nodes and add the following block into the Nomad Autoscaler configuration. 

If the `aws_access_key_id` and `aws_secret_access_key` settings are omitted the plugin will use the instance role to authenticate with the CloudWatch API. 

The IAM user or role requires the `cloudwatch:GetMetricData` permission.

```hcl
apm "cloudwatch" {
  driver = "nomad-autoscaler-cloudwatch-apm"

  config = {
    aws_region            = "us-east-1"
    aws_access_key_id     = "<AWS_ACCESS_KEY_ID>"
    aws_secret_access_key = "<AWS_SECRET_ACCESS_KEY>"
  }
}

```

## Policy Configuration

To scale a job with CloudWatch Metrics add the following block to your scaling policy. The query string is passed directly to the CloudWatch metrics API. Further details on the query syntax can be found in the [CloudWatch documentation](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/query_with_cloudwatch-metrics-insights.html).

```hcl
check "cloudwatch" {
  source = "cloudwatch"
  query  = <<-QUERY
    SELECT MAX(ApproximateNumberOfMessagesVisible) FROM SCHEMA("AWS/SQS", QueueName) WHERE QueueName = '<QUEUE_NAME>'
  QUERY

  strategy "target-value" {
    target = 50
  }
}
```

## Example

A Vagrant box with a working demo of the CloudWatch APM plugin has been provided in the [example](./example) folder.

