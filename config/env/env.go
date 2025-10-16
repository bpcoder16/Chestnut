package env

import (
	"time"

	"github.com/bpcoder16/Chestnut/v3/core/utils"
)

// 可以依据不同的运行等级来开启不同的调试功能、接口
const (
	// RunModeDebug 调试
	RunModeDebug = "debug"

	// RunModeTest 测试
	RunModeTest = "test"

	// RunModeRelease 线上发布
	RunModeRelease = "release"
)

// 基础配置类型
const (
	BaseConfigModeLocal = "local"

	BaseConfigModeNacos = "nacos"
)

// Option 具体的环境信息
//
// 所有的选项都是可选的
type Option struct {
	ConfigMode   string
	AppName      string
	RunMode      string
	TimeLocation string
}

// AppEnv 应用环境信息完整的接口定义
type AppEnv interface {
	RootPath() string
	LocalIPV4() string

	ConfigMode() string
	AppName() string
	RunMode() string
	TimeLocation() *time.Location
}

var _ AppEnv = (*appEnv)(nil)

type appEnv struct {
	rootPath  string
	localIPV4 string

	configMode   string
	appName      string
	runMode      string
	timeLocation *time.Location
}

func (a *appEnv) RootPath() string {
	return a.rootPath
}

func (a *appEnv) LocalIPV4() string {
	return a.localIPV4
}

func (a *appEnv) ConfigMode() string {
	if len(a.configMode) != 0 {
		return a.configMode
	}
	return BaseConfigModeLocal
}

func (a *appEnv) AppName() string {
	if len(a.appName) != 0 {
		return a.appName
	}
	return "unknown"
}

func (a *appEnv) RunMode() string {
	if len(a.runMode) != 0 {
		return a.runMode
	}
	return RunModeRelease
}

func (a *appEnv) TimeLocation() *time.Location {
	return a.timeLocation
}

func New(opt Option) AppEnv {
	env := &appEnv{}

	env.rootPath = utils.RootPath()

	var ipErr error
	env.localIPV4, ipErr = utils.GetLocalIPv4()
	if ipErr != nil {
		panic("GetLocalIPv4 err:" + ipErr.Error())
	}

	if len(opt.ConfigMode) != 0 {
		env.configMode = opt.ConfigMode
	}

	if len(opt.AppName) != 0 {
		env.appName = opt.AppName
	}

	if len(opt.RunMode) != 0 {
		env.runMode = opt.RunMode
	}

	if len(opt.TimeLocation) == 0 {
		opt.TimeLocation = "Asia/Shanghai"
	}
	var timeLocationErr error
	env.timeLocation, timeLocationErr = time.LoadLocation(opt.TimeLocation)
	if timeLocationErr != nil {
		panic("time.LoadLocation err:" + timeLocationErr.Error())
	}

	return env
}
