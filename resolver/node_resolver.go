package resolver

import (
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/ruimarinho/nsq-dogstatsd/collector"
	"github.com/ruimarinho/nsq-dogstatsd/internal/fetcher"
	"github.com/ruimarinho/nsq-dogstatsd/producer"
)

// ResolveNodes queries NSQD and NSQLookupd servers to retrieve information about producer nodes.
// Duplicate nodes are skipped, so if a NSQD address is passed and the same NSQD is found on a
// given NSQLookupd, it will only be used once.
func ResolveNodes(nsqdHTTPAddresses []string, lookupdHTTPAddresses []string) ([]producer.Producer, error) {
	var wg sync.WaitGroup
	var producerChan = make(chan producer.Producer)
	var errChan = make(chan error)

	for _, address := range nsqdHTTPAddresses {
		wg.Add(1)

		go func(address string) {
			defer wg.Done()

			fetcher := fetcher.NewFetcher(address)
			collector := collector.NSQDCollector{Fetcher: fetcher}
			info, err := collector.GetInfo()

			if err != nil {
				errChan <- err
				return
			}

			producerChan <- info.Producer
		}(address)
	}

	for _, address := range lookupdHTTPAddresses {
		wg.Add(1)

		go func(address string) {
			defer wg.Done()

			log.WithField("address", address).Print("resolving nodes from nsqlookupd")

			fetcher := fetcher.NewFetcher(address)
			collector := collector.NSQDCollector{Fetcher: fetcher}
			nodes, err := collector.GetNodes()

			if err != nil {
				errChan <- err
				return
			}

			for _, producer := range nodes.Data.Producers {
				producerChan <- producer
			}
		}(address)
	}

	go func() {
		wg.Wait()
		close(errChan)
		close(producerChan)
	}()

	addresses := map[string]bool{}
	producers := []producer.Producer{}

	for {
		select {
		case producer := <-producerChan:
			if producer.BroadcastAddress == "" {
				return producers, nil
			}

			if addresses[producer.HTTPAddress()] {
				log.WithField("address", producer.HTTPAddress()).Print("skipping duplicate address")
				continue
			}

			addresses[producer.HTTPAddress()] = true

			log.WithField("address", producer.HTTPAddress()).Print("added address")

			producers = append(producers, producer)
		case err := <-errChan:
			if err == nil {
				continue
			}

			return producers, err
		}
	}
}
