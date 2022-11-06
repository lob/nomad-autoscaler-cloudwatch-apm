## Testing

A Vagrant box with a Nomad has been provided for local testing.

1. First create an IAM user that has the `cloudwatch:GetMetricData` permssion. Add these credentials to `vagrant/jobs/autoscaler.nomad` jobspec.
2. Inspect the `vagrant/jobs/webapp.nomad` scaling configuration and update accordingly
3. Next compile the plugin by running `make dist` in the root folder.
4. Boot the VM by running `cd vagrant` and then `vagrant up`
5. Once the VM is setup run `cd jobs` and run the three jobs (haproxy first, then autoscaler, then webapp)
6. You can observe the logs for the autoscaler to see the CloudWatch API calls
