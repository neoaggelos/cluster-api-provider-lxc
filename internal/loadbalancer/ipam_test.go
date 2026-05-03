//go:build disabled

package loadbalancer

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"

	. "github.com/onsi/gomega"
)

func TestAllocate(t *testing.T) {
	g := NewWithT(t)

	opts, err := lxc.ConfigurationFromLocal("", "", false)
	g.Expect(err).ToNot(HaveOccurred())

	lxcClient, err := lxc.New(context.TODO(), opts)
	g.Expect(err).ToNot(HaveOccurred())

	as := make([]string, 30)
	errs := make([]error, 30)
	wg := sync.WaitGroup{}
	for i := range 30 {
		wg.Add(1)
		go func(i int) {
			as[i], errs[i] = (&ipam{
				lxcClient: lxcClient,

				networkName: "testbr0",

				rangesKey:   "user.capn.vip.ranges",
				volatileKey: func(s string) string { return fmt.Sprintf("user.capn.vip.volatile.%s", s) },
			}).Allocate(context.TODO(), fmt.Sprintf("default/c-%d", i))
			wg.Done()
		}(i)
	}

	wg.Wait()

	for i := range 30 {
		fmt.Println(as[i], errs[i])
	}
}
