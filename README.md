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

## Releasing a New Version

Releases are fully manual. There is no automated versioning — the version lives only in the GitHub Release tag.

### Steps

1. **Update CHANGELOG.md** — move entries from `[Unreleased]` into a new versioned section (e.g. `## [0.1.3] — YYYY-MM-DD`).

2. **Merge to `main`** — open a PR with your CHANGELOG update and any other changes, then merge it.

3. **Create a GitHub Release** — via the GitHub UI or CLI:
   ```bash
   gh release create v0.1.3 --title "v0.1.3" --notes "See CHANGELOG.md"
   ```
   Use a `v`-prefixed semver tag (e.g. `v0.1.3`). Creating the release triggers the `Release` workflow, which builds binaries for all platforms and attaches them to the release.

4. **Publish to S3** — once the `Release` workflow completes, run the `Publish to S3` workflow manually:
   - GitHub UI: **Actions → Publish to S3 → Run workflow**, enter the version tag (e.g. `v0.1.3`).
   - CLI: `gh workflow run publish-to-s3.yml -f version=v0.1.3`

   This uploads the four release binaries to:
   ```
   s3://lob-nomad-autoscaler-plugins/cloudwatch/<version>/nomad-autoscaler-cloudwatch-apm_linux_amd64
   s3://lob-nomad-autoscaler-plugins/cloudwatch/<version>/nomad-autoscaler-cloudwatch-apm_linux_arm64
   s3://lob-nomad-autoscaler-plugins/cloudwatch/<version>/nomad-autoscaler-cloudwatch-apm_darwin_amd64
   s3://lob-nomad-autoscaler-plugins/cloudwatch/<version>/nomad-autoscaler-cloudwatch-apm_darwin_arm64
   ```

5. **Update `terraform-services`** — bump the `cloudwatch_apm_plugin_version` variable in each environment where the plugin is deployed. The variable is set in the following files in the [`terraform-services`](https://github.com/lob/terraform-services) repo:

   | Environment | File |
   |---|---|
   | sandbox | `sandbox/nomad_stack/nomad/main.tf` |
   | staging | `staging/nomad_stack/nomad/main.tf` |
   | staging (render) | `staging/render_nomad_stack/nomad/main.tf` |
   | production | `production/nomad_stack/nomad/main.tf` |
   | production (render) | `production/render_nomad_stack/nomad/main.tf` |

   Change the value in each file, for example:
   ```hcl
   cloudwatch_apm_plugin_version = "v0.1.3"
   ```

   Open a PR, merge it, and apply the Terraform changes for each environment. The Nomad autoscaler will download the new plugin binary from S3 on its next restart or re-deploy.

