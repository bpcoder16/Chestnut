package utils

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bpcoder16/Chestnut/v4/logit"
)

func TimeCostLog(ctx context.Context, logField string) func() {
	start := time.Now()
	return func() {
		logit.Context(ctx).InfoW(logField+"_"+RandIntStr(3)+"_cost", strconv.FormatFloat(float64(time.Since(start).Nanoseconds())/1e6, 'f', 3, 64)+"ms")
	}
}

func ShowDurationString(duration time.Duration) string {
	return fmt.Sprintf("%.3fms", float64(duration.Nanoseconds())/1e6)
}

// ConvertTimeByTimezone
// timeStr 支持格式 time.RFC3339 / ISO 8601
// targetTZ 支持格式 GMT+08:00 / UTC+08:00
func ConvertTimeByTimezone(timeStr string, targetLocation *time.Location) (string, error) {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return "", fmt.Errorf("解析时间失败: %v", err)
	}

	targetTime := t.In(targetLocation)
	return targetTime.Format(time.DateTime), nil
}

// ParseOffsetToLocation
// targetTZ 支持格式 GMT+08:00 / UTC+08:00
func ParseOffsetToLocation(targetZone string) (*time.Location, error) {
	targetZoneNew := strings.TrimPrefix(targetZone, "GMT")
	targetZoneNew = strings.TrimPrefix(targetZoneNew, "UTC")
	sign := 1
	if strings.HasPrefix(targetZoneNew, "-") {
		sign = -1
		targetZoneNew = strings.TrimPrefix(targetZoneNew, "-")
	} else if strings.HasPrefix(targetZoneNew, "+") {
		targetZoneNew = strings.TrimPrefix(targetZoneNew, "+")
	} else {
		return nil, fmt.Errorf("传入的目标时区错误1: %s", targetZone)
	}

	parts := strings.Split(targetZoneNew, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("传入的目标时区错误2: %s", targetZone)
	}

	hours, errH := strconv.Atoi(parts[0])
	if errH != nil {
		return nil, errH
	}
	minutes, errM := strconv.Atoi(parts[1])
	if errM != nil {
		return nil, errM
	}

	totalSeconds := sign * (hours*60*60 + minutes*60)

	return time.FixedZone(targetZone, totalSeconds), nil
}
