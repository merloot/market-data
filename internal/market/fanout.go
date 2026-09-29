package market

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

func fanOut[K comparable, V any, R any](
	ctx context.Context,
	grouped map[K][]V,
	fn func(ctx context.Context, key K, items []V) (map[string]R, error),
) (map[string]R, error) {
	var (
		mu     sync.Mutex
		result = make(map[string]R)
		errs   []error
	)

	g, ctx := errgroup.WithContext(ctx)

	for key, items := range grouped {
		key, items := key, items
		g.Go(func() error {
			data, err := fn(ctx, key, items)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errs = append(errs, fmt.Errorf("%v: %w", key, err))
				return nil
			}

			for k, v := range data {
				result[k] = v
			}
			return nil
		})
	}
	_ = g.Wait()

	if len(errs) > 0 {
		return result, errors.Join(errs...)
	}
	return result, nil
}
