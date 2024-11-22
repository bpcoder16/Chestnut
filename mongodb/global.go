package mongodb

import (
	"context"
	"github.com/bpcoder16/Chestnut/contrib/gomongodb"
	"github.com/bpcoder16/Chestnut/core/log"
	"go.mongodb.org/mongo-driver/mongo"
)

var defaultManager *gomongodb.Manager

func SetManager(ctx context.Context, configPath string, logger *log.Helper) {
	defaultManager = gomongodb.NewManager(ctx, configPath, logger)
}

func DefaultClient() *mongo.Database {
	return defaultManager.ClientDatabase()
}
