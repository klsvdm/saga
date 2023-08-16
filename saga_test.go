package saga_test

import (
	"context"
	"github.com/klsvdm/saga"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	stepOneResult = "one"
	stepTwoResult = "two"
)

type testStep struct {
	result string
}

func (s *testStep) Exec(_ context.Context, data string) (string, error) {
	return data + s.result, nil
}

func (s *testStep) Revoke(_ context.Context) error {
	return nil
}

func TestTwoStepsSaga(t *testing.T) {
	const expected = stepOneResult + stepTwoResult

	testSaga := saga.New[string]()
	saga.AddStep[string, string](testSaga, &testStep{stepOneResult})
	saga.AddStep[string, string](testSaga, &testStep{stepTwoResult})

	err := testSaga.Run(context.Background())
	assert.Nil(t, err)

	result, err := testSaga.Result()
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}
