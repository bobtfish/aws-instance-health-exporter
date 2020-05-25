# AWS Instance Health Exporter

This is a simple exporter that scrapes the [describe-instance-status](https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-instance-status.html) API and exports it via HTTP for Prometheus consumption. That allows you to alert on pending scheduled events or hardware issues for your Aws instances


### Build
```
make
```

### Run
```
./aws-instance-health-exporter --aws.region=eu-west-1
```

## Exposed metrics
The `aws-instance-health-exporter` exports metrics for AWS instance events which are upcoming. The metric is a gauge of the number of seconds until the event.

Example
```
aws_instance_health_instance_events{event_code="instance-stop",instance_id="i-073d75d9dbede0a9a"} 650479.801696224
```

Name | Description | Labels
-----|-----|-----
aws_instance_health_instance_events | Upcoming AWS Health events for your instances | event_code, instance_id

### Labels Explained
Label | Description
-----|-----
event_code | The type of the event. Possible values are open, closed, and upcoming. For more info see the offical documentation [here](https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-instance-status.html)
instance_id | The EC2 instance ID to which the event applies.

## Flags
Flag | Description
-----|-----
`--help` | Show help.
`--version` | Print version information
`--web.listen-address` | The address to listen on for HTTP requests. Default: ":9165"
`--aws.region` | A list of AWS regions that are used to filter events

## Docker
You can deploy this exporter using the [bobtfish/aws-instance-health-exporter](https://hub.docker.com/r/bobtfish/aws-instance-health-exporter/) Docker Image.

Example
```
docker pull bobtfish/aws-instance-health-exporter
docker run -p 9165:9165 bobtfish/aws-instance-health-exporter
```

### Credentials
The `aws-instance-health-exporter` requires AWS credentials to access the AWS Health API. For example you can pass them via env vars using `-e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}` options, or you can suuply them in the environment via EC2 instance roles or ECS container roles.


