package pinger

import (
	"github.com/go-logr/logr"
	"github.com/launchboxio/agent/pkg/client"
	"time"
)

type Pinger struct {
	Client  *client.Client
	Logger  logr.Logger
	payload client.ClusterPing
}

func New(httpClient *client.Client, logger logr.Logger) *Pinger {
	return &Pinger{
		Client: httpClient,
		Logger: logger,
	}
}

func (p *Pinger) Init() error {
	// TODO: These should be sourced from the environment
	p.payload = client.ClusterPing{
		Version:         "1.25.15",
		AgentVersion:    "1.2.3",
		Provider:        "launchbox",
		Region:          "us-east-1",
		AgentIdentifier: "localhost",
	}
	return nil
}

func (p *Pinger) Start(clusterId int, interval time.Duration) {
	ticker := time.NewTicker(interval)
	quit := make(chan struct{})
	operator := p.Client.Operator()

	data := &client.PingRequest{
		Cluster: p.payload,
	}
	if _, err := operator.Ping(clusterId, data); err != nil {
		p.Logger.Error(err, "Failed to ping HQ")
	}
	go func() {
		for {
			select {
			case <-ticker.C:
				if _, err := operator.Ping(clusterId, data); err != nil {
					p.Logger.Error(err, "Failed to ping HQ")
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
