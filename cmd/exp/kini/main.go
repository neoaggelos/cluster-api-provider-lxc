package main

import (
	"context"
	"os"
	"path/filepath"

	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/lxc/cluster-api-provider-incus/cmd/exp/kini/docker"
	"github.com/lxc/cluster-api-provider-incus/cmd/exp/kini/kind"
	"github.com/lxc/cluster-api-provider-incus/cmd/exp/kini/kini"
)

var (
	ctx context.Context

	cmds = map[string]func(context.Context) error{
		"kini":   kini.NewCmd().ExecuteContext,
		"docker": docker.NewCmd().ExecuteContext,
		"kind":   kind.NewCmd().ExecuteContext,
	}

	log = ctrl.Log
)

func main() {
	defer klog.Flush()
	run, ok := cmds[filepath.Base(os.Args[0])]
	if !ok {
		run = cmds["kini"]
	}

	if err := run(ctx); err != nil {
		log.Error(err, "command failed")
		klog.Flush() // ensure error is flushed before exit
		os.Exit(1)
	}
}

func init() {
	ctx = ctrl.SetupSignalHandler()
	ctrl.SetLogger(klog.Background())
	ctx = ctrl.LoggerInto(ctx, log)
}
