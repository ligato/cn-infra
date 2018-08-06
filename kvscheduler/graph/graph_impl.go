package graph

import (
	"sync"
)

// kvgraph implements Graph interface.
type kvgraph struct {
	rwLock sync.RWMutex
	graph  *graphR
}

// NewGraph creates and new instance of key-value graph.
func NewGraph() Graph {
	kvgraph := &kvgraph{}
	kvgraph.graph = newGraphR()
	kvgraph.graph.parent = kvgraph
	return kvgraph
}

// Read returns a graph handle for read-only access.
// The graph supports multiple concurrent readers.
// Release eventually using Release() method.
func (kvgraph *kvgraph) Read() ReadAccess {
	kvgraph.rwLock.RLock()
	return kvgraph.graph
}

// Write returns a graph handle for read-write access.
// The graph supports at most one writer at a time - i.e. it is assumed
// there is no write-concurrency.
// The changes are propagated to the graph using Save().
// Release eventually using Release() method.
func (kvgraph *kvgraph) Write(record bool) RWAccess {
	return newGraphRW(kvgraph.graph, record)
}
