package graph

import (
	"time"

	. "github.com/ligato/cn-infra/kvscheduler/api"
	"github.com/ligato/cn-infra/idxmap"
)

// graphR implements GraphReadAccess.
type graphR struct {
	parent   *kvgraph
	nodes    map[string]*node
	mappings map[string]idxmap.NamedMappingRW
	timeline map[string][]*RecordedNode // key -> node records (from the oldest to the newest)
}

// newGraphR creates and initializes a new instance of graphR.
func newGraphR() *graphR {
	return &graphR{
		nodes:    make(map[string]*node),
		mappings: make(map[string]idxmap.NamedMappingRW),
		timeline: make(map[string][]*RecordedNode),
	}
}

// GetMetadataMap returns registered metadata map.
func (graph *graphR) GetMetadataMap(mapName string) idxmap.NamedMapping {
	metadataMap, has := graph.mappings[mapName]
	if !has {
		return nil
	}
	return metadataMap
}

// GetNode returns node with the given key or nil if the key is unused.
func (graph *graphR) GetNode(key string) Node {
	node, has := graph.nodes[key]
	if !has {
		return nil
	}
	return node.nodeR
}

// GetNodes returns a set of nodes matching the key selector (can be nil)
// and every provided flag selector.
func (graph *graphR) GetNodes(keySelector KeySelector, flagSelectors ...FlagSelector) (nodes []Node) {
	for key, node := range graph.nodes {
		if keySelector != nil && !keySelector(key) {
			continue
		}
		selected := true
		for _, flagSelector := range flagSelectors {
			for _, flag := range flagSelector.flags {
				hasFlag := false
				for _, nodeFlag := range node.flags {
					if nodeFlag.GetName() == flag.GetName() &&
						(flag.GetValue() == "" || (nodeFlag.GetValue() == flag.GetValue())) {
						hasFlag = true
						break
					}
				}
				if hasFlag != flagSelector.with {
					selected = false
					break
				}
			}
			if !selected {
				break
			}
		}
		if !selected {
			continue
		}
		nodes = append(nodes, node.nodeR)
	}
	return nodes
}

// GetNodeTimeline returns timeline of all node revisions, ordered from
// the oldest to the newest.
func (graph *graphR) GetNodeTimeline(key string) []*RecordedNode {
	timeline, has := graph.timeline[key]
	if !has {
		return nil
	}
	return timeline
}

// FlagStats returns stats for a given flag.
func (graph *graphR) FlagStats(flagName string, selector KeySelector) FlagStats {
	stats := FlagStats{PerValueCount: make(map[string]uint)}

	for key, timeline := range graph.timeline {
		if !selector(key) {
			continue
		}
		for _, record := range timeline {
			if flagValue, hasFlag := record.Flags[flagName]; hasFlag {
				stats.TotalCount++
				if _, hasValue := stats.PerValueCount[flagValue]; !hasValue {
					stats.PerValueCount[flagValue] = 0
				}
				stats.PerValueCount[flagValue]++
			}
		}
	}

	return stats
}

// GetSnapshot returns the snapshot of the graph at a given time.
func (graph *graphR) GetSnapshot(time time.Time) (nodes []*RecordedNode) {
	for _, timeline := range graph.timeline {
		for _, record := range timeline {
			if record.Since.Before(time) &&
				(record.Until.IsZero() || record.Until.After(time)) {
					nodes = append(nodes, record)
					break
			}
		}
	}
	return nodes
}

// Release releases the graph handle (both Read() & Write() should end with
// release).
func (graph *graphR) Release() {
	graph.parent.rwLock.RUnlock()
}

// copyNodesOnly returns a deep-copy of the graph, excluding the timelines
// and the map with mappings.
func (graph *graphR) copyNodesOnly() *graphR {
	graphCopy := &graphR{
		nodes: make(map[string]*node),
	}
	for key, node := range graph.nodes{
		nodeCopy := node.copy()
		nodeCopy.graph = graphCopy
		graphCopy.nodes[key] = newNode(nodeCopy)
	}
	return graphCopy
}

// recordNode builds a record for the node to be added into the timeline.
func (graph *graphR) recordNode(node *node) *RecordedNode {
	record := &RecordedNode{
		Since:          time.Now(),
		Key:            node.key,
		ValueLabel:     node.value.Label(),
		ValueType:      node.value.Type(),
		ValueString:    node.value.String(),
		Flags:          make(map[string]string),
		Targets:        node.targets, // no need to copy, never changed in graphR
	}
	for _, flag := range node.flags {
		record.Flags[flag.GetName()] = flag.GetValue()
	}
	if node.metaAdded {
		mapping := graph.mappings[node.metadataMap]
		record.MetadataFields = mapping.ListFields(node.value.Label())
	}

	return record
}