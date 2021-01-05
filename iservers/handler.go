package iservers

import (
	"context"
)

type Handler func(ctx context.Context, input map[string]interface{}) (interface{}, error)
