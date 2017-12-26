package app

import (
	"context"
)

type step interface {
	setup(context.Context, *App) error
	teardown(context.Context, *App) error
}

type steps []step

func (s steps) setup(ctx context.Context, app *App) error {
	for _, step := range s {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := step.setup(ctx, app)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s steps) teardown(ctx context.Context, app *App) []error {
	var errs []error

	for i := len(s) - 1; i >= 0; i-- {
		err := s[i].teardown(ctx, app)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func setupFunc(fn func(ctx context.Context, app *App) error) step {
	return &stepFn{fn: fn}
}

type stepFn struct {
	fn func(ctx context.Context, app *App) error
}

func (s *stepFn) setup(ctx context.Context, app *App) error {
	return s.fn(ctx, app)
}

func (s *stepFn) teardown(ctx context.Context, app *App) error {
	return nil
}
