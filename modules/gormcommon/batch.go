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

func LoopedReadData[T comparable](ctx context.Context, db *gorm.DB, where, order string, limit int) (result []T, err error) {
	page := 0
	result = make([]T, 0, limit*2)
	for {
		var resultTmp []T
		if err = db.WithContext(ctx).Where(where).Order(order).Limit(limit).Offset(page * limit).Find(&resultTmp).Error; err != nil || len(resultTmp) == 0 {
			return
		}
		result = append(result, resultTmp...)
		page++
	}
}
