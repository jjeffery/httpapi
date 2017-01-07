package httpapi_test

import (
	"context"
	"time"
)

type PostSomethingInput struct{}
type PostSomethingOutput struct{}

func postSomething(ctx context.Context, input *PostSomethingInput) (*PostSomethingOutput, error) {
	return nil, nil
}

type GetSomethingInput struct {
	Search string
	Since  time.Time
	Limit  int
	Offset int
}
type GetSomethingOutput struct{}

func getSomething(ctx context.Context, input *GetSomethingInput) (*GetSomethingOutput, error) {
	return nil, nil
}
