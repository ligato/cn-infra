package graph

import (
	. "github.com/ligato/cn-infra/kvscheduler/api"
)

type node struct {
	*nodeR

	metaAdded  bool
	metaInSync bool
	updated    bool
}

// newNode creates a new instance of node, either built from the scratch or
// extending existing nodeR.
func newNode(nodeR *nodeR) *node {
	if nodeR == nil {
		return &node{
			nodeR:      newNodeR(),
			metaInSync: true,
			updated:    true, /* completely new node */
		}
	}
	return &node{
		nodeR:      nodeR,
		metaInSync: true,
		updated:    false,
	}
}

// SetValue associates given value with this node.
func (node *node) SetValue(value Value) {
	node.value = value
	node.updated = true
}

// SetFlags associates given flag with this node.
func (node *node) SetFlags(flags ...Flag) {
	toBeSet := make(map[string]struct{})
	for _, flag := range flags {
		toBeSet[flag.GetName()] = struct{}{}
	}

	var otherFlags []Flag
	for _, flag := range node.flags {
		if _, set := toBeSet[flag.GetName()]; !set {
			otherFlags = append(otherFlags, flag)
		}
	}

	node.flags = append(otherFlags, flags...)
	node.updated = true
}

// DelFlags removes given flag from this node.
func (node *node) DelFlags(names ...string) {
	var otherFlags []Flag
	for _, flag := range node.flags {
		delete := false
		for _, flagName := range names {
			if flag.GetName() == flagName {
				delete = true
				break
			}
		}
		if !delete {
			otherFlags = append(otherFlags, flag)
		}
	}

	node.flags = otherFlags
	node.updated = true
}

// SetMetadataMap chooses metadata map to be used to store the association
// between this node's value label and metadata.
func (node *node) SetMetadataMap(mapName string) {
	if node.metadataMap == "" { // cannot be changed
		node.metadataMap = mapName
		node.updated = true
		node.metaInSync = false
	}
}

// SetMetadata associates given value metadata with this node.
func (node *node) SetMetadata(metadata interface{}) {
	node.metadata = metadata
	node.updated = true
	node.metaInSync = false
}

// SetTargets provides definition of all edges pointing from this node.
func (node *node) SetTargets(targets []RelationTarget) {
	node.targetsDef = targets
	node.updated = true

	// remove from sources of current targets
	node.removeThisFromSources()

	// re-init targets
	node.initRuntimeTarget()

	// build new targets
	for _, otherNode := range node.graph.nodes {
		if otherNode.key == node.key {
			continue
		}
		node.checkPotentialTarget(otherNode)
	}
}

// initRuntimeTarget re-initialize targets to empty key-sets.
func (node *node) initRuntimeTarget() {
	node.targets = make(map[string]RecordedTargets)

	for _, targetDef := range node.targetsDef {
		if _, hasRelation := node.targets[targetDef.Relation]; !hasRelation {
			node.targets[targetDef.Relation] = make(RecordedTargets)
		}
		if _, hasLabel := node.targets[targetDef.Relation][targetDef.Label]; !hasLabel {
			node.targets[targetDef.Relation][targetDef.Label] = make(KeySet)
		}
	}
}

// checkPotentialTarget checks if node2 is target of node in any of the relations.
func (node *node) checkPotentialTarget(node2 *node) {
	for _, targetDef := range node.targetsDef {
		if targetDef.Key == node2.key || (targetDef.Key == "" && targetDef.Selector(node2.key)) {
			node.targets[targetDef.Relation][targetDef.Label][node2.key] = struct{}{}
			node.updated = true
			node2.sources[targetDef.Relation][node.key] = struct{}{}
			node2.updated = true
		}
	}
}

// removeFromTargets removes given key from the map of targets.
func (node *node) removeFromTargets(key string) {
	for relation, targets := range node.targets {
		for label := range targets {
			if _, has := node.targets[relation][label][key]; has {
				delete(node.targets[relation][label], key)
				node.updated = true
			}
		}
	}
}

// removeFromTargets removes this node from the set of sources of all the other nodes.
func (node *node) removeThisFromSources() {
	for relation, targets := range node.targets {
		for _, targetNodes := range targets {
			for key := range targetNodes {
				targetNode := node.graph.nodes[key]
				delete(targetNode.sources[relation], node.GetKey())
				targetNode.updated = true
			}
		}
	}
}

// updateMetadataMap updates metadata in the associated mapping.
func (node *node) updateMetadataMap() {
	if !node.metaInSync {
		// update metadata map
		if mapping, hasMapping := node.graph.mappings[node.metadataMap]; hasMapping {
			if node.metadataAdded {
				mapping.Update(node.value.Label(), node.metadata)
			} else {
				mapping.Put(node.value.Label(), node.metadata)
			}
			node.metaAdded = true
		}
	}
}
