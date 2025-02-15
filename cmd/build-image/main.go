package main

import (
	"context"
	"flag"
	"os"

	"github.com/spf13/pflag"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	logsv1 "k8s.io/component-base/logs/api/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/neoaggelos/cluster-api-provider-lxc/internal/incus"
)

var (
	remoteName string
	configFile string

	imageBase string
	imageType string

	instanceName string
	instanceType string

	setupLog   = klog.Background().WithName("setup")
	logOptions = logs.NewOptions()

	ctx context.Context
)

func InitFlags(fs *pflag.FlagSet) {
	flag.StringVar(&remoteName, "remote", "", "Remote name to connect to")
	flag.StringVar(&configFile, "config", "", "Path to config file")
	flag.StringVar(&imageBase, "base", "ubuntu:24.04", "Base image. `ubuntu:XX.XX` uses `ubuntu:XX.XX` on LXD and `images:ubuntu/XX.XX/cloud` on Incus")
	flag.StringVar(&imageType, "type", "", "Type of image to build. One of kubeadm|haproxy")
	flag.StringVar(&instanceType, "instance-type", "container", "One of container|virtual-machine")
	flag.StringVar(&instanceName, "instance-name", "builder", "Name of builder instance")

	logsv1.AddFlags(logOptions, fs)
}

func main() {
	InitFlags(pflag.CommandLine)
	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	// Set log level 2 as default.
	if err := pflag.CommandLine.Set("v", "2"); err != nil {
		setupLog.Error(err, "Failed to set default log level")
		os.Exit(1)
	}
	pflag.Parse()

	if err := logsv1.ValidateAndApply(logOptions, nil); err != nil {
		setupLog.Error(err, "Unable to start manager")
		os.Exit(1)
	}
	ctx = klog.NewContext(ctrl.SetupSignalHandler(), klog.Background())

	options, err := incus.NewOptionsFromConfigFile(configFile, remoteName, false)
	if err != nil {
		log.FromContext(ctx).Error(err, "Failed to load server config file")
	}

	client, err := incus.New(ctx, options)
	if err != nil {
		log.FromContext(ctx).Error(err, "Failed to initialize incus client")
	}

	_ = client

}
