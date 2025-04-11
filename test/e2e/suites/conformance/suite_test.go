//go:build e2e

package conformance

import (
	"context"
	"testing"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/lxc/cluster-api-provider-incus/test/e2e/shared"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var e2eCtx *shared.E2EContext

func init() {
	e2eCtx = shared.NewE2EContext()
	shared.CreateDefaultFlags(e2eCtx)
}

func TestConformance(t *testing.T) {
	RegisterFailHandler(Fail)
	ctrl.SetLogger(GinkgoLogr)
	RunSpecs(t, "capn-conformance")
}

var _ = SynchronizedBeforeSuite(func(ctx context.Context) []byte {
	return shared.Node1BeforeSuite(ctx, e2eCtx)
}, func(ctx context.Context, data []byte) {
	shared.AllNodesBeforeSuite(e2eCtx, data)
})

var _ = SynchronizedAfterSuite(func(ctx context.Context) {
	shared.AllNodesAfterSuite(e2eCtx)
}, func(ctx context.Context) {
	shared.Node1AfterSuite(ctx, e2eCtx)
})
