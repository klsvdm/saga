package saga

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrAlreadyStarted     = errors.New("saga is already started")
	ErrCancelled          = errors.New("saga was cancelled")
	ErrUnexpectedDataType = errors.New("unexpected data type")
)

type Step[Data any, Result any] interface {
	Exec(ctx context.Context, data Data) (Result, error)
	Revoke(ctx context.Context) error
}

type (
	stepExecFunc   func(ctx context.Context, data any) (any, error)
	stepRevokeFunc func(ctx context.Context) error
)

type sagaStep struct {
	Exec   stepExecFunc
	Revoke stepRevokeFunc
}

type Saga[Result any] struct {
	steps   []sagaStep
	stepIdx int
	result  any

	canceled chan struct{}
}

func New[Result any]() *Saga[Result] {
	return &Saga[Result]{
		steps:    nil,
		stepIdx:  0,
		canceled: nil,
	}
}

func AddStep[Data, Result, R any](saga *Saga[R], step Step[Data, Result]) {
	if saga == nil || step == nil {
		return
	}

	saga.steps = append(saga.steps, sagaStep{
		Exec: func(ctx context.Context, data any) (any, error) {
			typedData, ok := data.(Data)
			if !ok {
				return nil, errors.Join(
					ErrUnexpectedDataType,
					fmt.Errorf("want '%T', got '%T'", typedData, data),
				)
			}

			return step.Exec(ctx, typedData)
		},
		Revoke: step.Revoke,
	})
}

func (s *Saga[Result]) Run(ctx context.Context, opts ...Option) (err error) {
	if s.canceled != nil {
		return ErrAlreadyStarted
	}

	ctx = context.WithValue(ctx, sagaKey{}, s)

	s.canceled = make(chan struct{})
	s.result = nil

	o := applyOptions(opts...)

	for s.stepIdx = 0; s.stepIdx < len(s.steps); s.stepIdx++ {
		if s.isCancelled() {
			return ErrCancelled
		}

		step := s.steps[s.stepIdx]

		s.result, err = runWithRetry(ctx, s.result, step.Exec, o.Retry)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Saga[Result]) Revoke(ctx context.Context) error {
	for idx := s.stepIdx; idx >= 0; idx-- {
		if err := s.steps[idx].Revoke(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *Saga[Result]) Cancel() {
	if s.canceled == nil {
		return
	}

	close(s.canceled)
}

func (s *Saga[Result]) Result() (Result, error) {
	var empty Result

	if s.isCancelled() {
		return empty, ErrCancelled
	}

	typedResult, ok := s.result.(Result)
	if !ok {
		return empty, errors.Join(
			ErrUnexpectedDataType,
			fmt.Errorf("want '%T', got '%T'", empty, s.result),
		)
	}

	return typedResult, nil
}

func (s *Saga[Result]) isCancelled() bool {
	select {
	case <-s.canceled:
		return true
	default:
	}

	return false
}
