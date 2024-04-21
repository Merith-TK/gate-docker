package dockerGate

import (
	"context"
	"fmt"
	"net" // Add missing import
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/go-logr/logr"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

var Plugin = proxy.Plugin{
	Name: "DockerGate",
	Init: func(ctx context.Context, p *proxy.Proxy) error {
		log := logr.FromContextOrDiscard(ctx)
		log.Info("Loading Docker")

		pl := &plugin{proxy: p}

		// register a loop to update the serverlist
		var wg sync.WaitGroup
		wg.Add(1)
		go updateServerList(ctx, pl, p, wg)

		return nil
	},
}

type plugin struct {
	proxy *proxy.Proxy
}

func updateServerList(ctx context.Context, pl *plugin, p *proxy.Proxy, wg sync.WaitGroup) {
	defer wg.Done()
	// Create a new Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		// Wait until Docker is running
		for {
			logr.FromContextOrDiscard(ctx).Error(err, "Failed to create Docker client, retrying in 1 second")
			cli, err = client.NewClientWithOpts(client.FromEnv)
			if err == nil {
				break
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		logr.FromContextOrDiscard(ctx).Info("Created Docker client")
	}
	// while the docker client is not nil
	for cli != nil {
		time.Sleep(1 * time.Second)
		// Use the Docker client to read the running containers
		containers, err := cli.ContainerList(context.Background(), container.ListOptions{
			Filters: filters.NewArgs(filters.KeyValuePair{
				Key:   "label",
				Value: "minekube.gate",
			}),
		})
		if err != nil {
			// Handle error
			return
		}

		// get list of existing servers
		servers := p.Servers()
		// logr.FromContextOrDiscard(ctx).Info("Updating server list", "servers", len(servers), "containers", len(containers))

		// Process the containers
		for _, container := range containers {
			containerName := container.Names[0][1:]

			serverFound := false
			for _, server := range servers {
				if server.ServerInfo().Name() == containerName {
					serverFound = true
					break
				}
			}
			if serverFound {
				continue
			}

			serverPort := container.Labels["minekube.port"]
			serverHost := container.Labels["minekube.host"]
			if serverHost == "" {
				logr.FromContextOrDiscard(ctx).Info("Server host not found, using container name", "container", containerName)
				serverHost = containerName
			}
			if serverPort == "" {
				logr.FromContextOrDiscard(ctx).Info("Server port not found, using default port 25565", "container", containerName)
				serverPort = "25565"
			}

			mergedHost := fmt.Sprintf("%s:%s", serverHost, serverPort)
			addrHost, err := net.ResolveTCPAddr("tcp", mergedHost)
			if err != nil {
				logr.FromContextOrDiscard(ctx).Error(err, "Failed to resolve address", "host", mergedHost)
				continue
			}
			serverInfo := proxy.NewServerInfo(containerName, addrHost)
			logr.FromContextOrDiscard(ctx).Info("Server info", "server", serverInfo)
			// check if the server already exists

			if !serverFound {
				logr.FromContextOrDiscard(ctx).Info("Registering new server", "server", containerName)
				p.Register(serverInfo)
				// TODO: Populate fallback list
			}
		}
	}
	logr.FromContextOrDiscard(ctx).Info("Docker client is nil, stopping server list update")
}
