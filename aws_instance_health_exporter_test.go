package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
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
	var statuses = []*ec2.InstanceStatus{}
	/*		&health.Event{
				EventTypeCategory: aws.String("issue"),
				Region:            aws.String("eu-west-1"),
				Service:           aws.String("EC2"),
				StatusCode:        aws.String("open"),
			},
			&health.Event{
				EventTypeCategory: aws.String("issue"),
				Region:            aws.String("us-east-1"),
				Service:           aws.String("EC2"),
				StatusCode:        aws.String("open"),
			},
			&health.Event{
				EventTypeCategory: aws.String("issue"),
				Region:            aws.String("us-east-1"),
				Service:           aws.String("LAMBDA"),
				StatusCode:        aws.String("closed"),
			},
			&health.Event{
				EventTypeCategory: aws.String("issue"),
				Region:            aws.String("us-east-1"),
				Service:           aws.String("LAMBDA"),
				StatusCode:        aws.String("closed"),
			},
			&health.Event{
				EventTypeCategory: aws.String("issue"),
				Region:            aws.String("us-east-1"),
				Service:           aws.String("LAMBDA"),
				StatusCode:        aws.String("closed"),
			},
		}*/
	e := &exporter{
		client: &mockEC2Client{statuses: statuses},
	}
	e.Collect(nil)
	//validateMetric(t, gv, events[0], 1.)
	//validateMetric(t, gv, events[1], 1.)
	//validateMetric(t, gv, events[2], 3.)
}

/*
func validateMetric(t *testing.T, vec *prometheus.GaugeVec, e *health.Event, expectedVal float64) {
	m := vec.WithLabelValues(*e.EventTypeCategory, *e.Region, *e.Service, *e.StatusCode)
	pb := &dto.Metric{}
	m.Write(pb)

	val := pb.GetGauge().GetValue()
	if pb.GetGauge().GetValue() != expectedVal {
		t.Errorf("Invalid value - Expected: %v Got: %v", expectedVal, val)
	}
} */
