package saga_test

import (
	"context"
	"fmt"
	"github.com/klsvdm/saga"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	stepOneResult = "one"
	stepTwoResult = "two"
)

type testStep struct {
	result  string
	revoked bool
}

func (s *testStep) Exec(_ context.Context, data string) (string, error) {
	return data + s.result, nil
}

func (s *testStep) Revoke(_ context.Context) error {
	s.revoked = true
	return nil
}

type testStepFailed struct {
	attempts int
	count    int
	revoked  bool
}

func (s *testStepFailed) Exec(_ context.Context, _ string) (int, error) {
	if s.count < s.attempts {
		s.count++
		return 0, fmt.Errorf("failed")
	}

	return 1, nil
}

func (s *testStepFailed) Revoke(_ context.Context) error {
	s.revoked = true
	return nil
}

func TestTwoStepsSaga(t *testing.T) {
	const expected = stepOneResult + stepTwoResult

	testSaga := saga.New[string]()
	saga.AddStep[string, string](testSaga, &testStep{result: stepOneResult})
	saga.AddStep[string, string](testSaga, &testStep{result: stepTwoResult})

	err := testSaga.Run(context.Background())
	assert.Nil(t, err)

	result, err := testSaga.Result()
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}

func TestStepRetry(t *testing.T) {
	const attempts = 5

	failedStep := &testStepFailed{attempts: attempts}

	testSaga := saga.New[int]()
	saga.AddStep[string, string](testSaga, &testStep{result: stepOneResult})
	saga.AddStep[string, int](testSaga, failedStep)

	err := testSaga.Run(context.Background(), saga.WithRetry(attempts+1, time.Microsecond))
	assert.Nil(t, err)

	assert.Equal(t, attempts, failedStep.count)
}

func TestSagaRevoke(t *testing.T) {
	const attempts = 5

	firstStep := &testStep{result: stepOneResult}
	failedStep := &testStepFailed{attempts: attempts + 1}

	testSaga := saga.New[int]()
	saga.AddStep[string, string](testSaga, firstStep)
	saga.AddStep[string, int](testSaga, failedStep)

	err := testSaga.Run(context.Background(), saga.WithRetry(attempts, time.Microsecond))
	assert.NotNil(t, err)

	err = testSaga.Revoke(context.Background())
	assert.Nil(t, err)

	assert.Equal(t, attempts, failedStep.count-1)
	assert.True(t, failedStep.revoked)
	assert.True(t, firstStep.revoked)
}
