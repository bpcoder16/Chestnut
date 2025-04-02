package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Service interface {
	Name(child Service) string
	Description() string
	Run(ctx context.Context, cmd *cobra.Command, args []string)
	SetFlags(flagSet *pflag.FlagSet)
}
