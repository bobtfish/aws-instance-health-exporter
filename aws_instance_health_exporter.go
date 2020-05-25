package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	Namespace = "aws_instance_health"
)

var (
	// BuildTime represents the time of the build
	BuildTime = "N/A"
	// Version represents the Build SHA-1 of the binary
	Version = "N/A"

	instanceEvents *prometheus.Desc
	statusOk       *prometheus.Desc
)

type exporter struct {
	client ec2iface.EC2API
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- instanceEvents
	ch <- statusOk
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	allIs := make([]*ec2.InstanceStatus, 0)
	more := true
	var nextToken *string
	for more == true {
		is, err := e.client.DescribeInstanceStatus(&ec2.DescribeInstanceStatusInput{
			IncludeAllInstances: aws.Bool(true),
			MaxResults:          aws.Int64(1000),
			NextToken:           nextToken,
		})
		if err != nil {
			panic(err)
		}
		for _, s := range is.InstanceStatuses {
			fmt.Println("instance %s instance state %s system status %s", *s.InstanceId, *s.InstanceStatus.Status, *s.SystemStatus.Status)
			allIs = append(allIs, s)
		}
		if is.NextToken == nil {
			more = false
		} else {
			nextToken = is.NextToken
		}
	}
	ch <- prometheus.MustNewConstMetric(instanceEvents, prometheus.GaugeValue, float64(1), "i-1234")
}

func init() {
	instanceEvents = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "health", "events"),
		"events for instances",
		[]string{"instance_id"},
		nil,
	)
	statusOk = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "health", "status_ok"),
		"Health status for instance",
		[]string{"health_type", "instance_id"},
		nil,
	)
	prometheus.MustRegister(version.NewCollector("aws_instance_health_exporter"))
}

func main() {
	var (
		showVersion = kingpin.Flag("version", "Print version information").Bool()
		listenAddr  = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9165").String()
		region      = kingpin.Flag("aws.region", "The AWS region").Default("us-east-1").String()
	)

	registerSignals()

	kingpin.Parse()

	if *showVersion {
		tw := tabwriter.NewWriter(os.Stdout, 2, 1, 2, ' ', 0)
		fmt.Fprintf(tw, "Build Time:   %s\n", BuildTime)
		fmt.Fprintf(tw, "Build SHA-1:  %s\n", Version)
		fmt.Fprintf(tw, "Go Version:   %s\n", runtime.Version())
		tw.Flush()
		os.Exit(0)
	}

	log.Printf("Starting `aws-instance-health-exporter`: Build Time: '%s' Build SHA-1: '%s'\n", BuildTime, Version)

	sess, err := session.NewSession(&aws.Config{Region: region})
	if err != nil {
		log.Fatal(err)
	}

	exporter := &exporter{client: ec2.New(sess)}
	prometheus.MustRegister(exporter)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>AWS Instance Health Exporter</title></head>
             <body>
             <h1>AWS Instance Health Exporter</h1>
             <p><a href='/metrics'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Println("Listening on", *listenAddr)
	http.ListenAndServe(*listenAddr, mux)
}

func registerSignals() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Print("Received SIGTERM, exiting...")
		os.Exit(1)
	}()
}
