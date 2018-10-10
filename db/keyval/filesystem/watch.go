package filesystem

import (
	"github.com/ligato/cn-infra/datasync"
)

// Implementation of BytesWatchResp (generic response)
type watchResp struct {
	op               datasync.Op
	key              string
	value, prevValue []byte
	rev              int64
}

func (wr *watchResp) GetValue() []byte {
	return wr.value
}

func (wr *watchResp) GetPrevValue() []byte {
	return wr.prevValue
}

func (wr *watchResp) GetKey() string {
	return wr.key
}

func (wr *watchResp) GetChangeType() datasync.Op {
	return wr.op
}

func (wr *watchResp) GetRevision() (rev int64) {
	return wr.rev
}
