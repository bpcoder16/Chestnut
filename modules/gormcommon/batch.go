package gormcommon

import (
	"context"

	"gorm.io/gorm"
)

func BatchInsertData[T comparable](ctx context.Context, db *gorm.DB, list []T, batchNum int) {
	for {
		if len(list) <= batchNum {
			if len(list) > 0 {
				db.WithContext(ctx).Create(&list)
			}
			break
		}
		data := list[:batchNum]
		db.WithContext(ctx).Create(&data)
		list = list[batchNum:]
	}
}

func LoopedReadData[T comparable](ctx context.Context, db *gorm.DB, result []T, order string, limit int, whereQuery any, whereArgs ...any) (resultNew []T, err error) {
	page := 0
	resultNew = make([]T, 0, len(result)*2)
	resultNew = append(resultNew, result...)
	for {
		var resultTmp []T
		if err = db.WithContext(ctx).Where(whereQuery, whereArgs...).Order(order).Limit(limit).Offset(page * limit).Find(&resultTmp).Error; err != nil || len(resultTmp) == 0 {
			return
		}
		resultNew = append(resultNew, resultTmp...)
		page++
	}
}
