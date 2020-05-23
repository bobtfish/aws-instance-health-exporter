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
The `aws-instance-health-exporter` exports metrics for instance status and instance events .

Example
```
# FIXME
aws_instance_events{status_code="open", region="us-east-1", category="issue"}
```

Name | Description | Labels
-----|-----|-----
# FIXME
aws_health_events | AWS Health events | category, region, service, status_code

### Labels Explained
Label | Description
-----|-----
# FIXMe
category | The category of the event. Possible events are issue, accountNotification and scheduledChange.
region | The AWS region name of the event. E.g. us-east-1.
service | The AWS service that is affected by the event. For example, EC2, RDS.
status_code | The most recent status of the event. Possible values are open, closed, and upcoming.

The labels match the corresponding `AWS Event` content - for a more detailed and up-to-date explanation see the offical documention [here](http://docs.aws.amazon.com/health/latest/APIReference/API_Event.html)

## Flags
Flag | Description
-----|-----
`--help` | Show help.
`--version` | Print version information
`--web.listen-address` | The address to listen on for HTTP requests. Default: ":9165"
`--aws.region` | A list of AWS regions that are used to filter events

## Docker
FIXME
You can deploy this exporter using the [bobtfish/aws-instance-health-exporter](https://hub.docker.com/r/bobtfish/aws-instance-health-exporter/) Docker Image.

Example
```
docker pull bobtfish/aws-instance-health-exporter
docker run -p 9165:9165 bobtfish/aws-instance-health-exporter
```

### Credentials
The `aws-instance-health-exporter` requires AWS credentials to access the AWS Health API. For example you can pass them via env vars using `-e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}` options, or you can suuply them in the environment via EC2 instance roles or ECS container roles.


