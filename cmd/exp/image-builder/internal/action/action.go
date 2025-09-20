package action

import "context"

type Action func(context.Context) error
