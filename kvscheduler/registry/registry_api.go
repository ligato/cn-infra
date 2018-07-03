package registry

import (
	. "github.com/ligato/cn-infra/kvscheduler/api"
)

// Registry can be used to register all descriptors and get quick (cached, O(log))
// lookups by keys.
type Registry interface {
	// RegisterDescriptor add new descriptor into the registry.
	RegisterDescriptor(descriptor KVDescriptor)

	// GetAllDescriptors returns all registered descriptors ordered by dump-dependencies.
	GetAllDescriptors() []KVDescriptor

	// GetDescriptor returns descriptor with the given name.
	GetDescriptor(name string) KVDescriptor

	// GetDescriptorForKey returns descriptor handling the given key.
	GetDescriptorForKey(key string) KVDescriptor
}