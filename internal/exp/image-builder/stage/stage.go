package stage

import "github.com/lxc/cluster-api-provider-incus/internal/exp/image-builder/action"

type Stage struct {
	Name   string
	Action action.Action
}
