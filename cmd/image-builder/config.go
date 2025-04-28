package main

import (
	"context"

	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/lxc/cluster-api-provider-incus/internal/incus"
)

var (
	gCtx       context.Context
	gLogger    = ctrl.Log
	logOptions = logs.NewOptions()

	// command-line arguments
	cfg struct {
		// client configuration
		configFile       string
		configRemoteName string

		// base image configuration
		ubuntuVersion string

		// builder configuration
		instanceName     string
		instanceProfiles []string
		instanceType     string

		// image alias configuration
		imageAlias string

		// output
		outputFile string
	}

	kubeadmCfg struct {
		kubernetesVersion string
		pullExtraImages   []string
	}

	// runtime configuration
	client *incus.Client
)

func init() {
	gCtx = ctrl.SetupSignalHandler()
	ctrl.SetLogger(klog.Background())
	gCtx = ctrl.LoggerInto(gCtx, gLogger)
}
