# CloudWatch Nomad Autoscaler Example

A Vagrant box with Nomad, Nomad Autoscaler and the CloudWatch APM driver for local testing.
## Getting Started

First create an IAM user that has the `cloudwatch:GetMetricData` permssion. Add these credentials to `example/jobs/autoscaler.nomad` jobspec.

Update the `example/jobs/webapp.nomad` scaling configuration accordingly

Next compile the plugin by running `make dist` in the root folder.

Once compiled you can boot the VM by running `cd example` and then `vagrant up`. This will start Nomad and the Nomad Autoscaler and the Demo app.

Nomad will be accessible on http://localhost:4646. Once running you will be able to observe the logs for the Autoscaler to see the CloudWatch API calls
