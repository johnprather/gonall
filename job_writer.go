package main

import "fmt"

// JobWriter is a writer for
type JobWriter struct {
	buf    []byte
	out    chan string
	server Server
}

// NewJobWriter returns a new JobWriter
func NewJobWriter(server Server, out chan string) *JobWriter {
	w := &JobWriter{}
	w.out = out
	w.server = server
	return w
}

// Write writes stuff
func (w *JobWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		switch b {
		case '\n':
			w.out <- fmt.Sprintf("%s: %s\n", w.server, string(w.buf))
			w.buf = make([]byte, 0)
		default:
			//fmt.Printf("%s: adding char: %c\n", w.server, b)
			w.buf = append(w.buf, b)
		}
	}
	return len(p), nil
}
