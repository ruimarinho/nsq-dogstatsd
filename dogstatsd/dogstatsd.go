package dogstatsd

import (
	"github.com/DataDog/datadog-go/v5/statsd"
	log "github.com/sirupsen/logrus"
)

// NewDogStatsDClient returns a preconfigured DogStatsD client with namespace and global tags.
func NewDogStatsDClient(dogstatsdAddress string, namespace string, tags []string) (*statsd.Client, error) {
	client, err := statsd.New(dogstatsdAddress, statsd.WithNamespace(namespace), statsd.WithTags(tags), statsd.WithoutTelemetry())
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{"namespace": namespace, "tags": tags}).Debug("configured dogstatsd client")

	return client, nil
}
