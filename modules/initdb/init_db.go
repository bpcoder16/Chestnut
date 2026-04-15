package initdb

import (
	"context"
	"errors"
	"path"
	"strconv"
	"strings"

	"github.com/bpcoder16/Chestnut/v4/appconfig/env"
	"github.com/bpcoder16/Chestnut/v4/core/file/operations"
	"gorm.io/gorm"
)

func Init(ctx context.Context, defaultVersion int, gormDB *gorm.DB, fileVersionSaveDir, migrateSQLFileDir string) {
	versionFilePath := path.Join(env.RootPath(), fileVersionSaveDir, "version")

	fileVersionValue, err := operations.ReadOrCreate[int](versionFilePath, defaultVersion)
	if err != nil {
		panic(err)
	}

	pendingSQLFiles := getPendingSQLFiles(migrateSQLFileDir, fileVersionValue)

	fileVersionValue++
	for {
		if sqlFileValue, isOK := pendingSQLFiles[fileVersionValue]; isOK {
			var sqlValue string
			if len(sqlFileValue.CreateOrAlter) > 0 {
				createOrAlterValue, errR := operations.ReadFile(
					path.Join(
						env.RootPath(),
						migrateSQLFileDir,
						sqlFileValue.CreateOrAlter,
					),
				)
				if errR != nil {
					panic(errR)
				}
				sqlValue += createOrAlterValue
			}
			if len(sqlFileValue.InsertOrUpdate) > 0 {
				insertOrUpdateValue, errR := operations.ReadFile(
					path.Join(
						env.RootPath(),
						migrateSQLFileDir,
						sqlFileValue.InsertOrUpdate,
					),
				)
				if errR != nil {
					panic(errR)
				}
				sqlValue += insertOrUpdateValue
			}
			sqlValueList := strings.Split(strings.Trim(sqlValue, " \n\t\r"), ";")
			errDB := gormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				for _, sql := range sqlValueList {
					if len(sql) > 0 {
						if errE := tx.Exec(sql).Error; errE != nil {
							return errors.New("SQL:" + sql + ", ERR:" + errE.Error())
						}
					}
				}
				return nil
			})
			if errDB != nil {
				panic(errDB)
			}
			_ = operations.WriteFile[int](versionFilePath, fileVersionValue)
		} else {
			break
		}
		fileVersionValue = fileVersionValue + 1
	}
}

type sqlFile struct {
	CreateOrAlter  string
	InsertOrUpdate string
}

func getPendingSQLFiles(migrateSQLFileDir string, fileVersionValue int) (pendingSQLFiles map[int]sqlFile) {
	sqlFilePaths, err := operations.ListFilesSorted(
		path.Join(
			env.RootPath(),
			migrateSQLFileDir,
		),
	)
	if err != nil {
		panic(err)
	}

	pendingSQLFiles = make(map[int]sqlFile, 10)
	for _, sqlFilePath := range sqlFilePaths {
		tmpArr := strings.Split(sqlFilePath[:len(sqlFilePath)-4], "_")
		if len(tmpArr) == 2 {
			indexInt, errS := strconv.Atoi(tmpArr[0])
			if errS != nil {
				panic(errS)
			}
			if indexInt > fileVersionValue {
				var sqlFileValue sqlFile
				var isOK bool
				if sqlFileValue, isOK = pendingSQLFiles[indexInt]; !isOK {
					sqlFileValue = sqlFile{}
				}

				switch tmpArr[1] {
				case "createOrAlter":
					sqlFileValue.CreateOrAlter = sqlFilePath
				case "insertOrUpdate":
					sqlFileValue.InsertOrUpdate = sqlFilePath
				}
				pendingSQLFiles[indexInt] = sqlFileValue
			}
		}
	}
	return
}
