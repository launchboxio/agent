package pinger

import (
	"github.com/go-logr/logr"
	"github.com/launchboxio/agent/pkg/evaluator"
	"github.com/launchboxio/launchbox-go-sdk/service/cluster"
	"time"
)

type Pinger struct {
	Client  *cluster.Client
	Logger  logr.Logger
	payload *cluster.UpdateClusterInput
}

func New(client *cluster.Client, logger logr.Logger) *Pinger {
	return &Pinger{
		Client: client,
		Logger: logger,
	}
}

func (p *Pinger) Init(clusterId int, evaluation *evaluator.Evaluation, version string) error {
	// TODO: These should be sourced from the environment
	p.payload = &cluster.UpdateClusterInput{
		Version:         evaluation.KubernetesVersion,
		AgentVersion:    version,
		Provider:        evaluation.Provider,
		Region:          evaluation.Region,
		AgentIdentifier: evaluation.AgentIdentifier,
		ClusterId:       clusterId,
	}
	return nil
}

func (p *Pinger) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	quit := make(chan struct{})

	if _, err := p.Client.Update(p.payload); err != nil {
		p.Logger.Error(err, "Failed to ping HQ")
	}
	go func() {
		for {
			select {
			case <-ticker.C:
				if _, err := p.Client.Update(p.payload); err != nil {
					p.Logger.Error(err, "Failed to ping HQ")
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
