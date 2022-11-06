# Nomad Autoscaler CloudWatch APM

A Nomad Autoscaler APM plugin to scale Nomad jobs using CloudWatch Metrics.

## Plugin Configuration

```hcl
apm "cloudwatch" {
  driver = "nomad-autoscaler-cloudwatch-apm"

  config = {
    aws_region            = "us-east-1"
    aws_access_key_id     = "CHANGEME"
    aws_secret_access_key = "CHANGEME"
    aws_session_token     = "CHANGEME"
  }
}

```


## Policy Configuration

```golang
scaling {
  enabled = true
  min     = 1
  max     = 20

  policy {
    cooldown = "20s"

    check "cloudwatch" {
      source = "cloudwatch"
      query  = <<-QUERY
        SELECT MAX(ApproximateNumberOfMessagesVisible) FROM SCHEMA("AWS/SQS", QueueName) WHERE QueueName = 'MY_QUEUE'
      QUERY

      strategy "target-value" {
        target = 50
      }
    }
  }
}
```

## Minimal IAM Policy

```
IAM POLICY
```





