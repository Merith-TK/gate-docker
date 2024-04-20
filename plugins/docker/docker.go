package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"go.minekube.com/gate/cmd/gate"
	jconfig "go.minekube.com/gate/pkg/edition/java/config"
	"gopkg.in/yaml.v3"

	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

var Plugin = proxy.Plugin{
	Name: "DockerGate",
	Init: func(ctx context.Context, p *proxy.Proxy) error {
		log := logr.FromContextOrDiscard(ctx)
		log.Info("Loading Docker")

		pl := &plugin{proxy: p}
		event.Subscribe(p

		return nil
	},
}

type plugin struct {
	proxy *proxy.Proxy
}


