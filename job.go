package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Job is an object representing an ssh job
type Job struct {
	command    string
	server     Server
	err        error
	output     string
	hostConfig ConfigBlock
	sshConfig  *ssh.ClientConfig
	client     *ssh.Client
	agent      ssh.AuthMethod
	agentConn  net.Conn
	clients    []*ssh.Client
}

// NewJob returns an instantiated Job object
func NewJob(server Server, command string) *Job {
	job := &Job{}
	job.server = server
	job.command = command
	job.hostConfig = config.forServer(server)
	job.getAgent()
	go job.run()
	return job
}

func (j *Job) run() {

	defer func() {
		for i := len(j.clients) - 1; i >= 0; i-- {
			j.clients[i].Close()
		}
		j.agentConn.Close()
		runner.doneCh <- j
	}()

	client, err := j.getClient(j.hostConfig)
	if err != nil {
		j.err = err
		return
	}

	session, err := client.NewSession()
	if err != nil {
		j.err = err
		return
	}
	defer session.Close()

	session.Stdout = NewJobWriter(j.server, runner.outCh)
	session.Stderr = NewJobWriter(j.server, runner.errCh)

	err = session.Run(j.command)
	if err != nil {
		runner.errCh <- fmt.Sprintf("%s: %s\n", j.server, err)
	}

	return
}

// getAgent initializes the ssh.ClientConfig object(s)
func (j *Job) getAgent() {
	conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err == nil {
		j.agentConn = conn
		j.agent = ssh.PublicKeysCallback(agent.NewClient(j.agentConn).Signers)
	}
}

func (j *Job) getClient(hostConfig ConfigBlock) (*ssh.Client, error) {
	sshConfig := &ssh.ClientConfig{
		User:    hostConfig.User,
		Auth:    []ssh.AuthMethod{j.agent},
		Timeout: hostConfig.Timeout,
	}
	if hostConfig.Host == j.hostConfig.Host && runner.password != "" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.Password(runner.password))
	}
	if hostConfig.ProxyHost != "" {
		proxyConfig := config.forServer(hostConfig.ProxyHost)
		proxyClient, err := j.getClient(proxyConfig)
		if err != nil {
			return nil, err
		}

		conn, err := proxyClient.Dial("tcp", hostConfig.hostPort())
		if err != nil {
			return nil, fmt.Errorf("%s: %s", hostConfig.hostPort(), err)
		}

		connection, chans, reqs, err := ssh.NewClientConn(conn, hostConfig.hostPort(), sshConfig)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", hostConfig.hostPort(), err)
		}

		client := ssh.NewClient(connection, chans, reqs)
		j.clients = append(j.clients, client)
		return client, nil
	}

	client, err := ssh.Dial("tcp", hostConfig.hostPort(), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", hostConfig.hostPort(), err)
	}

	j.clients = append(j.clients, client)
	return client, nil
}

// formatData formats output by prepending hostname to every line
func (j *Job) formatData(data string) (outStr string) {
	if len(data) == 0 {
		return
	}
	if data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}
	lines := strings.Split(data, "\n")
	var fLines []string
	for _, line := range lines {
		fLines = append(fLines, fmt.Sprintf("%s: %s", j.server, line))
	}
	outStr = strings.Join(fLines, "\n")
	if outStr != "" {
		outStr += "\n"
	}
	return outStr
}

// formatDataBytes returns byte array of formatData(data)
func (j *Job) formatDataBytes(data string) (outBytes []byte) {
	return []byte(j.formatData(data))
}
