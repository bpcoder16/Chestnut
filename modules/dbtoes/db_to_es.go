package dbtoes

import (
	"context"
	"errors"
	"github.com/bpcoder16/Chestnut/v2/contrib/esclientv7"
	"time"
)

type ESCommonItem interface {
	GetDocId() string
}

const limitNum = 2000

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
