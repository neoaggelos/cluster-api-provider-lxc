package stage

import "github.com/lxc/cluster-api-provider-incus/cmd/exp/image-builder/internal/action"

type Stage struct {
	Name   string
	Action action.Action
}
