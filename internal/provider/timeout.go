package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const defaultTimeout = 20 * time.Minute

type CreateFn = func(ctx context.Context, defaultTimeout time.Duration) (time.Duration, diag.Diagnostics)

func contextWithTimeout(ctx context.Context, timeoutFn CreateFn) (context.Context, context.CancelFunc, diag.Diagnostics) {
	timeout, diag := timeoutFn(ctx, defaultTimeout)
	if diag != nil {
		return ctx, nil, diag
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)

	return ctx, cancel, nil
}

type DataSourceWithTimeout struct {
}

type ResourceWithTimeout struct {
}

func (d *DataSourceWithTimeout) ContextWithTimeout(ctx context.Context, timeoutFn CreateFn) (context.Context, context.CancelFunc, diag.Diagnostics) {
	return contextWithTimeout(ctx, timeoutFn)
}

func (r *ResourceWithTimeout) ContextWithTimeout(ctx context.Context, timeoutFn CreateFn) (context.Context, context.CancelFunc, diag.Diagnostics) {
	return contextWithTimeout(ctx, timeoutFn)
}
