package cmd

import "context"

type Service interface {
	Name(child Service) string
	Description() string
	Run(ctx context.Context, args []string)
}
