package kvscheduler

import (
	"sort"
	"github.com/ligato/cn-infra/kvscheduler/graph"
)

// Order by operations (in average should yield the shortest sequence of operations):
//  1. modify with re-create
//  2. add
//  3. modify
//  4. delete
//
// Furthermore, adds and deletes are ordered by dependencies to limit temporary
// pending states.
func (scheduler *Scheduler) orderValuesByOp(graphR graph.GraphReadAccess, values []kvForTxn) []kvForTxn {
	var recreateVals, addVals, modifyVals, deleteVals []kvForTxn
	deps := make(map[string]keySet)

	for _, kv := range values {
		descriptor := scheduler.registry.GetDescriptorForKey(kv.key)
		// collect dependencies among changed values
		valDeps := descriptor.Dependencies(kv.key, kv.value)
		deps[kv.key] = make(keySet)
		for _, kv2 := range values {
			for _, dep := range valDeps {
				if kv2.key == dep.Key || (dep.AnyOf != nil && dep.AnyOf(kv2.key)) {
					deps[kv.key][kv2.key] = struct{}{}
				}
			}
		}

		if kv.value == nil {
			deleteVals = append(deleteVals, kv)
			continue
		}
		node := graphR.GetNode(kv.key)
		if node == nil || node.GetFlag(PendingFlagName) != nil {
			addVals = append(addVals, kv)
		}
		if descriptor.ModifyHasToRecreate(kv.key, node.GetValue(), kv.value, node.GetMetadata()) {
			recreateVals = append(recreateVals, kv)
		} else {
			modifyVals = append(modifyVals, kv)
		}
	}

	scheduler.orderValuesByDeps(addVals, deps, true)
	scheduler.orderValuesByDeps(deleteVals, deps, false)

	var ordered []kvForTxn
	ordered = append(ordered, recreateVals...)
	ordered = append(ordered, addVals...)
	ordered = append(ordered, modifyVals...)
	ordered = append(ordered, deleteVals...)
	return ordered
}

func (scheduler *Scheduler) orderValuesByDeps(values []kvForTxn, deps map[string]keySet, depFirst bool) {
	sort.Slice(values, func(i, j int) bool {
		if depFirst {
			return dependsOn(values[j].key, values[i].key, deps, len(values), 0)
		} else {
			return dependsOn(values[i].key, values[j].key, deps, len(values), 0)
		}
	})
}
