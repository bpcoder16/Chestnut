package gomongodb

import (
	"context"
	"fmt"
	"github.com/bpcoder16/Chestnut/logit"
	"go.mongodb.org/mongo-driver/event"
)

func startedMonitor(ctx context.Context, evt *event.CommandStartedEvent) {
	logit.Context(ctx).DebugW(
		"MongoDBMonitor", "Started",
		"Command", fmt.Sprintf("%v", evt.Command),
		"DatabaseName", evt.DatabaseName,
		"CommandName", evt.CommandName,
		"RequestID", evt.RequestID,
		"ConnectionID", evt.ConnectionID,
	)
}

func succeededMonitor(ctx context.Context, evt *event.CommandSucceededEvent) {
	logit.Context(ctx).DebugW(
		"MongoDBMonitor", "Succeeded",
		"CommandName", evt.CommandName,
		"RequestID", evt.RequestID,
		"ConnectionID", evt.ConnectionID,
		"Reply", fmt.Sprintf("%v", evt.Reply),
	)
}

func failedMonitor(ctx context.Context, evt *event.CommandFailedEvent) {
	logit.Context(ctx).WarnW(
		"MongoDBMonitor", "Succeeded",
		"CommandName", evt.CommandName,
		"RequestID", evt.RequestID,
		"ConnectionID", evt.ConnectionID,
		"Failure", evt.Failure,
	)
}
