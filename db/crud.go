package db

// PutDel enumerates 'Put' (meaning 'Create' or 'Update') and 'Delete' operations
type PutDel string

const (
	// Put represents a Create or Update operation
	Put PutDel = "Put"
	// Delete operation
	Delete PutDel = "Delete"
)
