package main

// ServerList is a slice of Servers
type ServerList []Server

// dup creates a copy of itself
func (sl ServerList) dup() ServerList {
	newList := make(ServerList, len(sl))
	copy(sl, newList)
	return newList
}

// add appends a server to the list
func (sl *ServerList) add(server Server) {
	*sl = append(*sl, server)
}

// next pops the oldest server off the slice and returns it or nil if empty
func (sl *ServerList) next() *Server {
	if len(*sl) == 0 {
		return nil
	}
	list := *sl
	next := list[0]
	rest := list[1:]
	*sl = rest
	return &next
}
