package saga

import (
	"context"
	"time"
)

func runWithRetry(ctx context.Context, data any, f stepExecFunc, r retry) (any, error) {
	attempts := 0

	for {
		result, err := f(ctx, data)
		if err != nil && attempts == r.Attempts {
			return nil, err
		}

		if err != nil {
			attempts++
			time.Sleep(r.Delay)

			continue
		}

		return result, nil
	}
}
