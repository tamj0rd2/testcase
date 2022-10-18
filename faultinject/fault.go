package faultinject

import (
	"context"
)

// Check is a fault-injection helper method which check if there is an injected fault(s) in the given context.
// It checks for errors injected as context value, or ensures to trigger a CallerFault.
// It is safe to use from production code.
func Check(ctx context.Context, faults ...any) error {
	if ctx == nil {
		return nil
	}
	for _, fault := range faults {
		if err, ok := ctx.Value(fault).(error); ok {
			return err
		}
	}
	if ic, ok := lookupInjectContext(ctx); ok {
		if err, ok := ic.check(); ok {
			<-ctx.Done()
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

// Finish is function that can be called from a deferred context,
// and will inject fault when a function finished its execution.
// The error pointer should point to the function's named return error variable.
// If the function encountered an actual error, fault injection is skipped.
// It is safe to use from production code.
func Finish(returnErr *error, ctx context.Context, faults ...any) {
	if ctx == nil {
		return
	}
	if *returnErr != nil {
		return
	}
	if err := Check(ctx, faults...); err != nil {
		*returnErr = err
	}
}