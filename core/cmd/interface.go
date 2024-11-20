package cmd

import "context"

type Service interface {
	Name() string
	Description() string
	Run(ctx context.Context, args []string)
}
