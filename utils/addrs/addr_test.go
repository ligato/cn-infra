package addrs

import (
	"github.com/onsi/gomega"
	"testing"
)

func TestMacIntToString(t *testing.T) {
	gomega.RegisterTestingT(t)
	res := MacIntToString(0)
	gomega.Expect(res).To(gomega.BeEquivalentTo("00:00:00:00:00:00"))

	res = MacIntToString(255)
	gomega.Expect(res).To(gomega.BeEquivalentTo("00:00:00:00:00:ff"))
}
