package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/prometheus/client_golang/prometheus"
	"testing"
	"time"
)

type mockEC2Client struct {
	ec2iface.EC2API
	statuses []*ec2.InstanceStatus
}

func (client *mockEC2Client) DescribeInstanceStatus(in *ec2.DescribeInstanceStatusInput) (*ec2.DescribeInstanceStatusOutput, error) {
	output := ec2.DescribeInstanceStatusOutput{
		InstanceStatuses: client.statuses,
		NextToken:        nil,
	}
	return &output, nil
}

func TestScrape(t *testing.T) {
	now := time.Now()
	var statuses = []*ec2.InstanceStatus{
		&ec2.InstanceStatus{
			InstanceId: aws.String("i-1234"),
			Events: []*ec2.InstanceStatusEvent{
				&ec2.InstanceStatusEvent{
					Code:        aws.String("foo"),
					NotBefore:   &now,
					Description: aws.String("something reboot this way comes"),
				},
			},
		},
		&ec2.InstanceStatus{
			InstanceId: aws.String("i-2345"),
			Events: []*ec2.InstanceStatusEvent{
				&ec2.InstanceStatusEvent{
					Code:        aws.String("foo"),
					NotBefore:   &now,
					Description: aws.String("[Completed] something reboot this way comes"),
				},
			},
		},
	}
	e := &exporter{
		client: &mockEC2Client{statuses: statuses},
	}
	c := make(chan prometheus.Metric)
	go func() {
		defer close(c)
		e.Collect(c)
	}()
	metrics := make([]prometheus.Metric, 0)
	for m := range c {
		metrics = append(metrics, m)
		fmt.Println(m)
	}

	expMetrics := 1

	if len(metrics) != expMetrics {
		t.Errorf("Got %d metrics, not %d", len(metrics), expMetrics)
	}
}
