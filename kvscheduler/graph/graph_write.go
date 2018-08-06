package graph

import (
	"github.com/ligato/cn-infra/idxmap"
	"time"
)

// graphRW implements RWAccess.
type graphRW struct {
	*graphR
	record  bool
	deleted []string
	newRevs KeySet
}

// newGraphRW creates a new instance of grapRW, which extends an existing
// graph with write-operations.
func newGraphRW(graph *graphR, recordChanges bool) *graphRW {
	graphRCopy := graph.copyNodesOnly()
	return &graphRW{
		graphR:  graphRCopy,
		record:  recordChanges,
		newRevs: make(KeySet),
	}
}

// RegisterMetadataMap registers new metadata map for value-label->metadata
// associations of selected node.
func (graph *graphRW) RegisterMetadataMap(mapName string, mapping idxmap.NamedMappingRW) {
	if graph.mappings == nil {
		graph.mappings = make(map[string]idxmap.NamedMappingRW)
	}
	graph.mappings[mapName] = mapping
}

// SetNode creates new node or returns read-write handle to an existing node.
// The changes are propagated to the graph only after Save() is called.
// If <newRev> is true, the changes will recorded as a new revision of the
// node for the history.
func (graph *graphRW) SetNode(key string) NodeRW {
	node, has := graph.nodes[key]
	if has {
		return node
	}
	node = newNode(nil)
	node.key = key
	for _, otherNode := range graph.nodes {
		otherNode.checkPotentialTarget(node)
	}
	graph.nodes[key] = node

	return node
}

// DeleteNode deletes node with the given key.
// Returns true if the node really existed before the operation.
func (graph *graphRW) DeleteNode(key string) bool {
	node, has := graph.nodes[key]
	if !has {
		return false
	}

	// remove from sources of current targets
	node.removeThisFromSources()

	// delete from graph
	delete(graph.nodes, key)

	// remove from targets of other nodes
	for _, otherNode := range graph.nodes {
		otherNode.removeFromTargets(key)
	}
	graph.deleted = append(graph.deleted, key)
	return true
}

// Save propagates all changes to the graph.
func (graph *graphRW) Save() {
	graph.parent.rwLock.Lock()
	defer graph.parent.rwLock.Unlock()

	destGraph := graph.parent.graph

	// propagate newly registered mappings
	for mapName, mapping := range graph.mappings {
		if _, alreadyReg := destGraph.mappings[mapName]; !alreadyReg {
			destGraph.mappings[mapName] = mapping
		}
	}

	// apply deleted nodes
	for _, key := range graph.deleted {
		if node, has := destGraph.nodes[key]; has {
			node.metadata = nil
			node.updateMetadataMap()
			delete(destGraph.nodes, key)
		}
		graph.newRevs[key] = struct{}{}
	}
	graph.deleted = []string{}

	// apply new/changes nodes
	for key, node := range graph.nodes {
		if !node.updated {
			continue
		}
		node.graph = destGraph // move from working space to the actual graph
		destGraph.nodes[key] = node
		node.updateMetadataMap()
		graph.newRevs[key] = struct{}{}
	}

	// copy moved nodes
	for key, node := range graph.nodes {
		if !node.updated {
			continue
		}
		nodeCopy := node.copy()
		nodeCopy.graph = graph.graphR
		graph.nodes[key] = newNode(nodeCopy)
	}
}

// Release records changes if requested.
func (graph *graphRW) Release() {
	if graph.record {
		destGraph := graph.parent.graph
		for key := range graph.newRevs {
			node, exists := destGraph.nodes[key]
			if _, hasTimeline := destGraph.timeline[key]; !hasTimeline {
				if !exists {
					// deleted, but never recorded => skip
					continue
				}
				destGraph.timeline[key] = []*RecordedNode{}
			}
			records := destGraph.timeline[key]
			if len(records) > 0 {
				records[len(records)-1].Until = time.Now()
			}
			if exists {
				destGraph.timeline[key] = append(records, destGraph.recordNode(node))
			}
		}
	}
}
