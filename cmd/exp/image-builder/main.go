package main

import (
	"context"
	"os"

	"github.com/lxc/cluster-api-provider-incus/cmd/exp/image-builder/internal/cmd"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	ctx context.Context
	log = ctrl.Log
)

func main() {
	if err := cmd.NewCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	ctx = ctrl.SetupSignalHandler()
	ctrl.SetLogger(klog.Background())
	ctx = ctrl.LoggerInto(ctx, log)
}
