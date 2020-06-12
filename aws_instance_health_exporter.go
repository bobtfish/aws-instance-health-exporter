package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
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
)

type instanceEvent struct {
	instanceId string
	code       string
	eventTime  time.Time
}

func getEvents(client ec2iface.EC2API) ([]instanceEvent, error) {
	more := true
	var nextToken *string
	events := make([]instanceEvent, 0)
	for more == true {
		is, err := client.DescribeInstanceStatus(&ec2.DescribeInstanceStatusInput{
			IncludeAllInstances: aws.Bool(false),
			MaxResults:          aws.Int64(1000),
			NextToken:           nextToken,
		})
		if err != nil {
			return events, err
		}
		for _, s := range is.InstanceStatuses {
			eventCount := len(s.Events)
			if eventCount > 0 {
				for _, e := range s.Events {
					if strings.HasPrefix(*e.Description, "[Completed]") {
						continue
					}
					event := instanceEvent{
						eventTime:  *e.NotBefore,
						code:       *e.Code,
						instanceId: *s.InstanceId,
					}
					events = append(events, event)
				}
			}
		}
		if is.NextToken == nil {
			more = false
		} else {
			nextToken = is.NextToken
		}
	}
	return events, nil
}

var (
	cachedEvents []instanceEvent
	cacheMutex   sync.Mutex
	cachedAt     time.Time
)

// Deliberately return the cached result for concurrent calls, even
// if cache time is set to zero to avoid hitting the API serially
func getEventsCached(client ec2iface.EC2API, howOld time.Duration) ([]instanceEvent, error) {
	now := time.Now() // Store now from before we try to get the lock
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	if cachedAt.Add(howOld).Before(now) {
		events, err := getEvents(client)
		if err != nil { // Do not populate the cache if there was an error
			return events, err
		}
		cachedEvents = events
		// Set cachedAt to now *after* getting events
		cachedAt = time.Now()
	}
	return cachedEvents, nil
}

type exporter struct {
	client   ec2iface.EC2API
	cacheFor time.Duration
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- instanceEvents
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	events, err := getEventsCached(e.client, e.cacheFor)
	if err != nil {
		panic(err)
	}
	for _, e := range events {
		d := e.eventTime.Sub(time.Now())
		ch <- prometheus.MustNewConstMetric(instanceEvents, prometheus.GaugeValue, d.Seconds(), e.code, e.instanceId)
	}
}

func init() {
	cachedEvents = make([]instanceEvent, 0)
	instanceEvents = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "instance", "events"),
		"Upcoming events for instances - metric value is the number of seconds until the event",
		[]string{"event_code", "instance_id"},
		nil,
	)
	prometheus.MustRegister(version.NewCollector("aws_instance_health_exporter"))
}

func main() {
	var (
		showVersion            = kingpin.Flag("version", "Print version information").Bool()
		listenAddr             = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9165").String()
		region                 = kingpin.Flag("aws.region", "The AWS region").Default("us-east-1").String()
		cache                  = kingpin.Flag("cache", "The amount of time to cache results for. (Prometheus time format, e.g. 10s, 5m)").Default("0").String()
		disableExporterMetrics = kingpin.Flag(
			"web.disable-exporter-metrics",
			"Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).",
		).Bool()
		exporterMetricsRegistry = prometheus.NewRegistry()
		metricsRegistry         = prometheus.NewRegistry()
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

	if !*disableExporterMetrics {
		exporterMetricsRegistry.MustRegister(
			prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
			prometheus.NewGoCollector(),
		)
	}

	sess, err := session.NewSession(&aws.Config{Region: region})
	if err != nil {
		log.Fatal(err)
	}

	cacheDuration, err := model.ParseDuration(*cache)
	if err != nil {
		log.Fatal(err)
	}

	exporter := &exporter{
		client:   ec2.New(sess),
		cacheFor: time.Duration(cacheDuration),
	}
	metricsRegistry.MustRegister(exporter)
	handler := promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{})
	if !*disableExporterMetrics {
		handler = promhttp.InstrumentMetricHandler(exporterMetricsRegistry, handler)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", handler)
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
