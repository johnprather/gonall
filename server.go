package main

import (
	"errors"
	"regexp"
)

// Server is a hostname:port string
type Server string

var (
	// ErrServerInvalid is thrown by NewServer for invalid server names
	ErrServerInvalid = errors.New("invalid server name")
)

var serverRe *regexp.Regexp

func init() {
	serverRe = regexp.MustCompile(`(?i)^[0-9a-z\-]+(\.[0-9a-z\-]+)*(:[0-9]+)?$`)
}

// NewServer returns a new Server object or an error
func NewServer(name string) (Server, error) {
	if serverRe.MatchString(name) {
		return Server(name), nil
	}
	return "", ErrServerInvalid
}
