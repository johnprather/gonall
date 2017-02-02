package main

import (
	"flag"
	"fmt"
	"os"
)

// Flags is a struct of commandline args
type Flags struct {
	delay   int64
	jobs    int
	pass    bool
	user    string
	timeout int
}

var flags Flags
var args []string

func init() {
	flag.BoolVar(&flags.pass, "p", false, "prompt for a password to use")
	flag.IntVar(&flags.jobs, "n", 6, "number parallel ssh jobs")
	flag.Int64Var(&flags.delay, "d", 10, "delay (ms) between job starts")
	flag.StringVar(&flags.user, "u", "", "username to use")
	flag.IntVar(&flags.timeout, "t", 0, "timeout (secs) for connect to server")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n\n"+
			"  %s [options] command < hostfile\n\n"+
			"  ... | %s [options] <command>\n\n"+
			"Description:\n\n Options:\n\n",
			os.Args[0], os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
	}

	flag.Parse()
	args = flag.Args()

}
