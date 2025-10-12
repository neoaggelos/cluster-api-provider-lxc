package action

import (
	"context"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

type Action func(context.Context, *lxc.Client) error
