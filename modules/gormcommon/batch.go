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
