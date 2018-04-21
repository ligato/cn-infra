package consul

import (
	"testing"
	"time"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	. "github.com/onsi/gomega"
)

func init() {
	logrus.DefaultLogger().SetLevel(logging.DebugLevel)
}

type testCtx struct {
	store *Store
}

func setupTest(t *testing.T) *testCtx {
	RegisterTestingT(t)

	store, err := NewConsulStore("127.0.0.1:8500")
	if err != nil {
		t.Fatal("connecting to consul failed:", err)
	}

	return &testCtx{store}
}

func (ctx *testCtx) teardownTest() {
	ctx.store.Close()
}

func TestPut(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	err := ctx.store.Put("key", []byte("val"))
	Expect(err).ToNot(HaveOccurred())
}

func TestGetValue(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	data, found, rev, err := ctx.store.GetValue("key")
	Expect(err).ToNot(HaveOccurred())
	Expect(data).To(Equal([]byte("val")))
	Expect(found).To(BeTrue())
	Expect(rev).NotTo(BeZero())
}

func TestDelete(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	existed, err := ctx.store.Delete("key")
	Expect(err).ToNot(HaveOccurred())
	Expect(existed).To(BeTrue())
}

func TestWatch(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	closeCh := make(chan string)
	go func() {
		time.Sleep(time.Second)
		err := ctx.store.Put("key", []byte("val"))
		Expect(err).ToNot(HaveOccurred())
		close(closeCh)
	}()

	var resp = func(resp keyval.BytesWatchResp) {
		Expect(resp.GetChangeType()).To(Equal(datasync.Put))
		Expect(resp.GetValue()).To(Equal([]byte("val")))
	}
	err := ctx.store.Watch(resp, closeCh, "key")
	Expect(err).ToNot(HaveOccurred())
}
