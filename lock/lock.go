package lock

import (
	"context"
	"github.com/bpcoder16/Chestnut/v2/logit"
	"github.com/bpcoder16/Chestnut/v2/modules/lock/local"
)

var LocalManager *local.LockManager

func InitLocalManager(size int) {
	LocalManager = local.NewLocalLockManager(size)
}

func CleanupLockManager(ctx context.Context) {
	beforeCnt := LocalManager.Len()
	if beforeCnt > 0 {
		LocalManager.Cleanup()
		afterCnt := LocalManager.Len()
		logit.Context(ctx).InfoW("CleanupLocalLockManager", "success", "beforeCnt", beforeCnt, "afterCnt", afterCnt)
	} else {
		logit.Context(ctx).InfoW("CleanupLocalLockManager", "NoNeedToClean", "beforeCnt", beforeCnt)
	}
}
