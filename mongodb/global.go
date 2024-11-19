package mongodb

import (
	"context"
	"github.com/bpcoder16/Chestnut/contrib/gomongodb"
	"github.com/bpcoder16/Chestnut/core/log"
	"go.mongodb.org/mongo-driver/mongo"
)

var defaultMongodbManager *gomongodb.Manager

func SetManager(ctx context.Context, configPath string, logger *log.Helper) {
	defaultMongodbManager = gomongodb.NewManager(ctx, configPath, logger)
}

func DefaultClient() *mongo.Database {
	return defaultMongodbManager.ClientDatabase()
}
