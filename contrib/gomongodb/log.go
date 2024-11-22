package gomongodb

import (
	"context"
	"fmt"
	"github.com/bpcoder16/Chestnut/core/log"
	"go.mongodb.org/mongo-driver/event"
)

func startedMonitorFunc(l *log.Helper) func(context.Context, *event.CommandStartedEvent) {
	return func(ctx context.Context, evt *event.CommandStartedEvent) {
		l.WithContext(ctx).DebugW(
			"MongoDBMonitor", "Started",
			"Command", fmt.Sprintf("%v", evt.Command),
			"DatabaseName", evt.DatabaseName,
			"CommandName", evt.CommandName,
			"RequestID", evt.RequestID,
			"ConnectionID", evt.ConnectionID,
		)
	}
}

func succeededMonitorFunc(l *log.Helper) func(context.Context, *event.CommandSucceededEvent) {
	return func(ctx context.Context, evt *event.CommandSucceededEvent) {
		l.WithContext(ctx).DebugW(
			"MongoDBMonitor", "Succeeded",
			"CommandName", evt.CommandName,
			"RequestID", evt.RequestID,
			"ConnectionID", evt.ConnectionID,
			"Reply", fmt.Sprintf("%v", evt.Reply),
		)
	}
}

func failedMonitorFunc(l *log.Helper) func(ctx context.Context, evt *event.CommandFailedEvent) {
	return func(ctx context.Context, evt *event.CommandFailedEvent) {
		l.WithContext(ctx).WarnW(
			"MongoDBMonitor", "Succeeded",
			"CommandName", evt.CommandName,
			"RequestID", evt.RequestID,
			"ConnectionID", evt.ConnectionID,
			"Failure", evt.Failure,
		)
	}
}
