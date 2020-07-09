package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
)

var (
	host       string
	tls        bool
	tlscacert  string
	tlscert    string
	tlskey     string
	tlsverify  bool
	help       bool
	maxNameLen int
)

func main() {
	home, err := homedir.Dir()
	capath := ""
	certpath := ""
	keypath := ""
	if err == nil {
		capath = filepath.Join(home, ".docker", "ca.pem")
		certpath = filepath.Join(home, ".docker", "cert.pem")
		keypath = filepath.Join(home, ".docker", "key.pem")
	}

	app := cli.NewApp()
	app.HideHelp = true
	app.Name = "ifc"
	app.Usage = "interface configuration of docker container"
	app.Version = "0.1.0"
	app.ArgsUsage = ""
	app.UsageText = "ifc [global options]"
	app.Copyright = "Copyright (c) 2020 tenfy"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{Name: "tenfy", Email: "tenfy@tenfy.cn"},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "host, H",
			Value:       "unix:///var/run/docker.sock",
			Usage:       "Daemon `socket` to connect to",
			Destination: &host,
		},
		cli.BoolFlag{
			Name:        "tls",
			Usage:       "Use TLS; implied by --tlsverify",
			Destination: &tls,
		},
		cli.StringFlag{
			Name:        "tlscacert",
			Usage:       "Trust certs signed only by this CA",
			Value:       capath,
			Destination: &tlscacert,
		},
		cli.StringFlag{
			Name:        "tlscert",
			Usage:       "Path to TLS certificate file",
			Value:       certpath,
			Destination: &tlscert,
		},
		cli.StringFlag{
			Name:        "tlskey",
			Usage:       "Path to TLS key file",
			Value:       keypath,
			Destination: &tlskey,
		},
		cli.BoolFlag{
			Name:        "tlsverify",
			Usage:       "Use TLS and verity the remote",
			Destination: &tlsverify,
		},
		cli.IntFlag{
			Name:        "max_name_len",
			Usage:       "Max name len",
			Value:       20,
			Destination: &maxNameLen,
		},
		cli.BoolFlag{
			Name:        "help, h",
			Usage:       "Show help",
			Destination: &help,
		},
	}

	if maxNameLen == 0 {
		maxNameLen = 20
	}

	app.Action = func(c *cli.Context) error {
		if help {
			cli.ShowAppHelpAndExit(c, 0)
		}
		client, err := newClient()
		if err != nil {
			return err
		}

		ifc(client)
		return nil
	}

	app.Run(os.Args)
}

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
