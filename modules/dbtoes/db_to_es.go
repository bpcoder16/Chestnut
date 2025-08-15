package dbtoes

import (
	"context"
	"errors"
	"time"

	"github.com/bpcoder16/Chestnut/v2/contrib/esclientv7"
)

type ESCommonItem interface {
	GetDocId() string
}

const limitNum = 2000
const defaultRetryTimes = 3

func TimePeriodRetry(ctx context.Context, esManager *esclientv7.Manager, index string, lastTime time.Time, lastId uint64, esDataListFunc func(lastTime time.Time, lastId uint64, limit int) (time.Time, uint64, []any, error), retryTimes uint8) error {
	if retryTimes == 0 {
		retryTimes = defaultRetryTimes
	}
	var dataList []any
	var err error
	var realRetryTimes uint8
	for {
		lastTime, lastId, dataList, err = esDataListFunc(lastTime, lastId, limitNum)
		if err != nil {
			if realRetryTimes >= retryTimes {
				return err
			}
			retryTimes++
			continue
		}
		if len(dataList) == 0 {
			break
		}
		documentList := make([]esclientv7.Document, 0, len(dataList))
		for _, v := range dataList {
			esData := v
			if esCommonItem, ok := esData.(ESCommonItem); ok {
				documentList = append(documentList, esclientv7.Document{
					ID:      esCommonItem.GetDocId(),
					Content: esData,
				})
			}
		}
		if len(documentList) == 0 {
			return errors.New("documentList.Empty")
		}
		if errES := esManager.BulkUpsert(ctx, index, documentList); errES != nil {
			return errES
		}
		realRetryTimes = 0
	}

	return nil
}

func TimePeriod(ctx context.Context, esManager *esclientv7.Manager, index string, lastTime time.Time, lastId uint64, esDataListFunc func(lastTime time.Time, lastId uint64, limit int) (time.Time, uint64, []any, error)) error {
	var dataList []any
	var err error
	for {
		lastTime, lastId, dataList, err = esDataListFunc(lastTime, lastId, limitNum)
		if err != nil {
			return err
		}
		if len(dataList) == 0 {
			break
		}
		documentList := make([]esclientv7.Document, 0, len(dataList))
		for _, v := range dataList {
			esData := v
			if esCommonItem, ok := esData.(ESCommonItem); ok {
				documentList = append(documentList, esclientv7.Document{
					ID:      esCommonItem.GetDocId(),
					Content: esData,
				})
			}
		}
		if len(documentList) == 0 {
			return errors.New("documentList.Empty")
		}
		if errES := esManager.BulkUpsert(ctx, index, documentList); errES != nil {
			return errES
		}
	}

	return nil
}

func Default(ctx context.Context, esManager *esclientv7.Manager, index string, lastId uint64, esDataListFunc func(lastId uint64, limit int) (uint64, []any, error)) error {
	var dataList []any
	var err error
	for {
		lastId, dataList, err = esDataListFunc(lastId, limitNum)
		if err != nil {
			return err
		}
		if len(dataList) == 0 {
			break
		}
		documentList := make([]esclientv7.Document, 0, len(dataList))
		for _, v := range dataList {
			esData := v
			if esCommonItem, ok := esData.(ESCommonItem); ok {
				documentList = append(documentList, esclientv7.Document{
					ID:      esCommonItem.GetDocId(),
					Content: esData,
				})
			}
		}
		if len(documentList) == 0 {
			return errors.New("documentList.Empty")
		}
		if errES := esManager.BulkUpsert(ctx, index, documentList); errES != nil {
			return errES
		}
	}

	return nil
}
