package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

// Runner struct for the object that handles ... everything ...
type Runner struct {
	serverCh    chan Server
	command     string
	doneCh      chan *Job
	limit       int
	inputDoneCh chan bool
	password    string
	outCh       chan string
	errCh       chan string
}

var runner *Runner

func init() {
	runner = NewRunner()
}

// NewRunner returns an instantiated Runner
func NewRunner() *Runner {
	r := &Runner{}
	r.serverCh = make(chan Server)
	r.doneCh = make(chan *Job)
	r.inputDoneCh = make(chan bool)
	r.outCh = make(chan string)
	r.errCh = make(chan string)
	return r
}

func (r *Runner) start() {
	r.limit = flags.jobs
	r.getPasswords()
	r.getCommand()
	go r.listenServerList()
	r.listen()
}

// getCommand preps r.command
func (r *Runner) getCommand() {
	switch len(args) {
	case 0:
		flag.Usage()
		os.Exit(2)
	case 1:
		r.command = args[0]
	default:
		r.command = ""
		for _, arg := range args {
			r.command = fmt.Sprintf("%s \"%s\"", r.command,
				strings.Replace(arg, "\"", "\\\"", -1))
		}
	}
}

// getPasswords sets passwords if necessary
func (r *Runner) getPasswords() {
	if flags.pass {
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0666)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer tty.Close()
		fd := tty.Fd()
		tty.Write([]byte("Enter password for ssh target hosts: "))
		pass, err := terminal.ReadPassword(int(fd))
		tty.Write([]byte("\n"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		r.password = string(pass)
	}
}

// listen is our main data synced reader/writer for thread-safe ops
func (r *Runner) listen() {
	var servers ServerList
	var running int
	var inputDone bool
	var isDone = func() bool {
		if inputDone && running == 0 && len(servers) == 0 {
			return true
		}
		return false
	}

	for {
		if isDone() {
			return
		}
		select {
		case server := <-r.serverCh:
			servers.add(server)
		case job := <-r.doneCh:
			r.handleJobResults(job)
			running--
		case <-r.inputDoneCh:
			inputDone = true
		case str := <-r.outCh:
			os.Stdout.Write([]byte(str))
		case str := <-r.errCh:
			os.Stderr.Write([]byte(str))
		case <-time.After(time.Duration(flags.delay) * time.Millisecond):
			if running < r.limit {
				if running < r.limit {
					if next := servers.next(); next != nil {
						running++
						NewJob(*next, r.command)
					}
				}
			}
		}
	}
}

// handleJobResults quickly handles completed jobs
func (r *Runner) handleJobResults(job *Job) {
	/*
		if job.output != "" {
			os.Stdout.Write(job.formatDataBytes(job.output))
		}
		if job.err != nil {
			os.Stderr.Write(job.formatDataBytes(job.err.Error()))
		}
	*/
}

// listenServerList reads on the passed reader for a server list until err (EOF)
func (r *Runner) listenServerList() {
	var buf []byte
	b := make([]byte, 1)
	for _, err := os.Stdin.Read(b); err == nil; _, err = os.Stdin.Read(b) {
		switch b[0] {
		case '\n':
			if server, err := NewServer(strings.TrimSpace(string(buf))); err == nil {
				r.serverCh <- server
			}
			buf = make([]byte, 0)
		default:
			buf = append(buf, b[0])
		}
	}
	r.inputDoneCh <- true
}
