package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lxc/cluster-api-provider-incus/internal/static"
)

type stageInstallHaproxy struct{}

func (*stageInstallHaproxy) name() string { return "install-haproxy" }

// cat embed/50-install-haproxy.sh | incus exec capn-builder -- bash -s
// incus config set capn-builder user.capn.stage.install-haproxy=true
func (*stageInstallHaproxy) run(ctx context.Context) error {
	instance, etag, err := lxcClient.GetInstance(cfg.instanceName)
	if err != nil {
		return fmt.Errorf("failed to retrieve instance info: %w", err)
	}

	if instance.Config["user.capn.stage.install-haproxy"] == "true" {
		log.FromContext(ctx).V(1).Info("Stage already run")
		return nil
	}

	stdin := bytes.NewBufferString(static.InstallHaproxyScript())

	var stdout, stderr io.Writer
	if log.FromContext(ctx).V(4).Enabled() {
		stdout = os.Stdout
		stderr = os.Stderr
	}

	log.FromContext(ctx).V(1).Info("Running install-haproxy.sh script")
	if err := lxcClient.RunCommand(ctx, cfg.instanceName, []string{"bash", "-s"}, stdin, stdout, stderr); err != nil {
		return fmt.Errorf("failed to run install-haproxy.sh script: %w", err)
	}

	log.FromContext(ctx).V(1).Info("Set user.capn.stage.install-haproxy=true on instance")
	instance.Config["user.capn.stage.install-haproxy"] = "true"
	if _, err := lxcClient.UpdateInstance(cfg.instanceName, instance.InstancePut, etag); err != nil {
		log.FromContext(ctx).V(1).Info("WARNING: Failed to set user.capn.stage.install-haproxy=true on instance", "error", err)
	}

	return nil
}
