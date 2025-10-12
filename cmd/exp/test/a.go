package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	opts, _, err := lxc.ConfigurationFromLocal("", "", false)
	if err != nil {
		panic(err)
	}

	klog.InitFlags(nil)
	flag.Set("v", "5")
	ctrl.SetLogger(klog.Background())

	ctrl.Log.Info("Loaded", "opts", opts)

	client, err := lxc.New(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	profiles, err := client.GetProfileNames()
	if err != nil {
		panic(err)
	}

	fmt.Println(profiles)
}
