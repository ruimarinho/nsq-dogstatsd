package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/ruimarinho/nsq-dogstatsd/collector"
	"github.com/ruimarinho/nsq-dogstatsd/dogstatsd"
	"github.com/ruimarinho/nsq-dogstatsd/internal/checker"
	"github.com/ruimarinho/nsq-dogstatsd/internal/parser"
	"github.com/ruimarinho/nsq-dogstatsd/internal/slice"
	"github.com/ruimarinho/nsq-dogstatsd/producer"
	"github.com/ruimarinho/nsq-dogstatsd/resolver"
	log "github.com/sirupsen/logrus"
)

var (
	interval                = flag.Duration("interval", time.Duration(0), `interval for collecting metrics (default "none")`)
	namespace               = flag.String("namespace", "nsq", "namespace for metrics")
	dogstatsdAddress        = flag.String("dogstatsd-address", "127.0.0.1:8125", "<address>:<port> to connect to dogstatsd")
	showVersion             = flag.Bool("version", false, "show version information")
	stagger                 = flag.Bool("stagger", false, "spread the collection of nsqd nodes evenly across -interval, instead of collecting all nodes at once (delays the first collection of each node by its offset)")
	excludeMetricsPatterns  slice.StringSlice
	nsqdHTTPAddresses       slice.StringSlice
	nsqlookupdHTTPAddresses slice.StringSlice
	tags                    slice.StringSlice
	verbose                 = flag.Int("verbose", 0, "verbosity level (0-3)")
	version                 = "master"
)

func init() {
	flag.Var(&excludeMetricsPatterns, "exclude-metrics", "exclude metrics using a regular expression pattern (can be specified multiple times)")
	flag.Var(&tags, "tag", `add global tags (can be specified multiple times)`)
	flag.Var(&nsqdHTTPAddresses, "nsqd-http-address", "<address>:<port> of nsqd node to query stats for (can be specified multiple times)")
	flag.Var(&nsqlookupdHTTPAddresses, "lookupd-http-address", "<address>:<port> of nsqlookupd to query nodes for (can be specified multiple times)")
}

// staggerOffset returns the offset of the i-th of n nodes within the interval,
// so that the nodes are spaced evenly.
func staggerOffset(i int, n int, interval time.Duration) time.Duration {
	return interval * time.Duration(i) / time.Duration(n)
}

func collectNode(p producer.Producer, client *statsd.Client, excludeMetrics []*regexp.Regexp) error {
	c := collector.NewCollector(p, excludeMetrics)
	metrics, err := c.CollectMetrics()
	if err != nil {
		return err
	}

	for _, m := range metrics {
		if err = client.Gauge(m.Name, m.Value, m.Tags, m.Rate); err != nil {
			return err
		}
	}

	return client.Flush()
}

func sendMetricsLoop(nsqdHTTPAddresses []string, lookupdHTTPAddresses []string, dogstatsdAddress string, namespace string, tags []string, excludeMetrics []*regexp.Regexp, interval time.Duration, doneChan chan bool, errChan chan error) {
	producers, err := resolver.ResolveNodes(nsqdHTTPAddresses, lookupdHTTPAddresses)

	if err != nil {
		errChan <- err
		return
	}

	client, err := dogstatsd.NewDogStatsDClient(dogstatsdAddress, namespace, tags)
	if err != nil {
		errChan <- err
		return
	}

	log.WithField("interval", interval.String()).Info("interval set")

	if *stagger {
		sort.Slice(producers, func(i, j int) bool {
			return producers[i].HTTPAddress() < producers[j].HTTPAddress()
		})
	}

	// Each node is collected on its own ticker, so that a slow node does not
	// delay the others.
	var wg sync.WaitGroup
	for i, p := range producers {
		var offset time.Duration
		if *stagger {
			offset = staggerOffset(i, len(producers), interval)

			log.WithFields(log.Fields{"node": p.Hostname, "address": p.HTTPAddress(), "offset": offset.String()}).Info("staggering collection")
		}

		wg.Add(1)

		go func(p producer.Producer, offset time.Duration) {
			defer wg.Done()

			time.Sleep(offset)

			// Trigger initial metrics collection instead of waiting for first tick,
			// which could be far in the future. The ticker is created first so the
			// duration of this collection does not delay the following ones.
			var timeChan <-chan time.Time
			if interval.Seconds() > 0 {
				timeChan = time.NewTicker(interval).C
			}

			for {
				if err := collectNode(p, client, excludeMetrics); err != nil {
					errChan <- err
					return
				}

				if timeChan == nil {
					return
				}

				<-timeChan
			}
		}(p, offset)
	}

	wg.Wait()

	doneChan <- true
}

func main() {
	flag.Parse()

	if flag.NFlag() == 0 {
		fmt.Fprint(os.Stderr, "Usage of nsq_to_dogstatsd:\n\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	if *showVersion {
		fmt.Printf("nsq_to_dogstatsd %s\n", version)
		os.Exit(0)
	}

	if len(nsqdHTTPAddresses) == 0 && len(nsqlookupdHTTPAddresses) == 0 {
		log.Fatal("--lookup-http-address or --nsqd-http-address must be provided at least once")
	}

	if err := checker.CheckAddresses(nsqdHTTPAddresses); err != nil {
		log.Fatalf("--nsqd-http-address - %s", err)
	}

	if err := checker.CheckAddresses(nsqlookupdHTTPAddresses); err != nil {
		log.Fatalf("--nsqlookupd-http-address - %s", err)
	}

	if *stagger && *interval <= 0 {
		log.Fatal("--stagger requires a positive --interval")
	}

	excludedMetrics, err := parser.Parse(excludeMetricsPatterns)
	if err != nil {
		log.Fatalf("--filter-metrics contains invalid regexp - %s", err)
	}

	switch *verbose {
	case 0:
		log.SetLevel(log.ErrorLevel)
	case 1:
		log.SetLevel(log.WarnLevel)
	case 2:
		log.SetLevel(log.InfoLevel)
	case 3:
		log.SetLevel(log.DebugLevel)
	default:
		log.Fatalf("--verbose is outside valid range (0-3)")
	}

	doneChan := make(chan bool)
	errChan := make(chan error)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go sendMetricsLoop(nsqdHTTPAddresses, nsqlookupdHTTPAddresses, *dogstatsdAddress, *namespace, tags, excludedMetrics, *interval, doneChan, errChan)

	select {
	case <-doneChan:
		log.Info("exiting")
		os.Exit(0)
	case err := <-errChan:
		log.WithField("error", err).Fatal("exiting due to error")
	case signal := <-signalChan:
		log.WithField("signal", signal).Info("exiting due to signal")
		os.Exit(0)
	}
}
