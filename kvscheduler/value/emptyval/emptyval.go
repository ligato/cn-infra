package emptyval

import (
	. "github.com/ligato/cn-infra/kvscheduler/api"
)

// emptyValue can be used whenever the mere existence of the value is the only
// information needed (typically Property values).
type emptyValue struct {
	valueType ValueType
}

// NewEmptyValue creates a new instance of empty value.
func NewEmptyValue(valueType ValueType) Value {
	return &emptyValue{valueType:valueType}
}

// Label returns empty string.
func (ev *emptyValue) Label() string {
	return ""
}

// Equivalent returns true for two empty values.
func (ev *emptyValue) Equivalent(v2 Value) bool {
	_, isEmpty := v2.(*emptyValue)
	if !isEmpty {
		return false
	}
	return true
}

// String returns empty string.
func (ev * emptyValue) String() string {
	return ""
}

// Type returns the type selected in NewEmptyValue constructor.
func (ev * emptyValue) Type() ValueType {
	return ev.valueType
}