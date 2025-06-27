package dbtoes

import (
	"context"
	"errors"
	"github.com/bpcoder16/Chestnut/v2/contrib/esclientv7"
)

type ESCommonItem interface {
	GetDocId() string
}

const limitNum = 2000

func TimePeriod(ctx context.Context, esManager *esclientv7.Manager, index string, esDataListFunc func(limit, offset int) ([]any, error)) error {
	startOffset := 0
	for {
		dataList, err := esDataListFunc(limitNum, startOffset*limitNum)
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
		startOffset++
	}

	return nil
}
