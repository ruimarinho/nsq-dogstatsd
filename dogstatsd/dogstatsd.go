package dogstatsd

import (
	"fmt"

	"github.com/DataDog/datadog-go/statsd"
	log "github.com/sirupsen/logrus"
)

// NewDogStatsDClient returns a preconfigured DogStatsD client with namespace and global tags.
func NewDogStatsDClient(dogstatsdAddress string, namespace string, tags []string) (*statsd.Client, error) {
	client, err := statsd.New(dogstatsdAddress)
	if err != nil {
		return nil, err
	}

	client.Namespace = fmt.Sprintf("%s.", namespace)
	client.Tags = tags

	log.WithFields(log.Fields{"namespace": namespace, "tags": tags}).Debug("configured dogstatsd client")

	return client, nil
}
