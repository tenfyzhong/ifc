package main

import (
	"fmt"

	docker "github.com/fsouza/go-dockerclient"
)

func newClient() (*docker.Client, error) {
	if tls {
		return docker.NewTLSClient(host, tlscert, tlskey, tlscacert)
	}
	return docker.NewClient(host)
}

func ifc(client *docker.Client) {
	if client == nil {
		return
	}

	containers, err := client.ListContainers(docker.ListContainersOptions{
		All: false,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%-16s%-8s%-17s%-19s%-17s%-16s%-16s%s\n", "ID", "Network", "IPv4", "Mac", "Gateway", "EndpointID", "NetworkID", "Name")
	for _, container := range containers {
		id := formatStr(container.ID, 12)
		name := ""
		if len(container.Names) > 0 {
			name = formatStr(container.Names[0][1:], maxNameLen)
		}
		for ntype, network := range container.Networks.Networks {
			endpointID := formatStr(network.EndpointID, 12)
			networkID := formatStr(network.NetworkID, 12)
			fmt.Printf(
				"%-16s%-8s%-17s%-19s%-17s%-16s%-16s%s\n",
				id,
				formatStr(ntype, 6),
				network.IPAddress,
				network.MacAddress,
				network.Gateway,
				endpointID,
				networkID,
				name)
		}
	}
}

func formatStr(str string, max int) string {
	if len(str) > max {
		return str[0:max]
	}
	return str
}
