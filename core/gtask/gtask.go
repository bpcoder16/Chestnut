package gtask

import (
	"context"
	"golang.org/x/sync/errgroup"
)

type Group struct {
	*errgroup.Group
}

func WithContext(ctx context.Context) (*Group, context.Context) {
	g := &Group{}
	g.Group, ctx = errgroup.WithContext(ctx)
	return g, ctx
}

func (g *Group) Go(f func() error) {
	g.Group.Go(f)
}

func (g *Group) Wait() error {
	return g.Group.Wait()
}
