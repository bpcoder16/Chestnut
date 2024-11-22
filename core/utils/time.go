package utils

import (
	"context"
	"github.com/bpcoder16/Chestnut/logit"
	"strconv"
	"time"
)

func TimeCostLog(ctx context.Context, logField string) func() {
	start := time.Now()
	return func() {
		logit.Context(ctx).InfoW(logField+"_"+RandIntStr(3)+"_cost", strconv.FormatFloat(float64(time.Since(start).Nanoseconds())/1e6, 'f', 3, 64)+"ms")
	}
}
