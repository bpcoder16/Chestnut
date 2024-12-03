package env

import (
	"github.com/bpcoder16/Chestnut/core/utils"
	"path"
	"time"
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

// Option 具体的环境信息
//
// 所有的选项都是可选的
type Option struct {
	// AppName 应用名称
	AppName string

	// RunMode 运行模式
	RunMode string

	TimeLocation string

	ConfigDirName string
}

// AppEnv 应用环境信息完整的接口定义
type AppEnv interface {
	// AppNameEnv 应用名称
	AppNameEnv

	// RunModeEnv 应用运行情况
	RunModeEnv

	RootPath() string

	LocalIPV4() string

	TimeLocation() *time.Location

	ConfigDirPath() string
}

// AppNameEnv 应用名称
type AppNameEnv interface {
	AppName() string
}

// RunModeEnv 运行模式/等级
type RunModeEnv interface {
	RunMode() string
}

var _ AppEnv = (*appEnv)(nil)

type appEnv struct {
	appName       string
	runMode       string
	rootPath      string
	configDirPath string
	localIPV4     string
	timeLocation  *time.Location
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

func (a *appEnv) RootPath() string {
	return a.rootPath
}

func (a *appEnv) ConfigDirPath() string {
	return a.configDirPath
}

func (a *appEnv) LocalIPV4() string {
	return a.localIPV4
}

func (a *appEnv) TimeLocation() *time.Location {
	return a.timeLocation
}

func (a *appEnv) setAppName(name string) {
	setValue(&a.appName, name, "AppName")
}

func (a *appEnv) setRunMode(mod string) {
	setValue(&a.runMode, mod, "RunMode")
}

func (a *appEnv) setConfigDirPath(path string) {
	setValue(&a.configDirPath, path, "ConfigDirPath")
}

func setValue(addr *string, value string, fieldName string) {
	*addr = value
}

func New(opt Option) AppEnv {
	env := &appEnv{}

	if len(opt.AppName) != 0 {
		env.setAppName(opt.AppName)
	}

	if len(opt.RunMode) != 0 {
		env.setRunMode(opt.RunMode)
	}

	env.rootPath = utils.RootPath()
	var ipErr error
	env.localIPV4, ipErr = getLocalIPv4()
	if ipErr != nil {
		panic(ipErr)
	}

	if len(opt.TimeLocation) == 0 {
		opt.TimeLocation = "Asia/Shanghai"
	}
	var timeLocationErr error
	env.timeLocation, timeLocationErr = time.LoadLocation(opt.TimeLocation)
	if timeLocationErr != nil {
		panic(timeLocationErr)
	}

	if len(opt.ConfigDirName) == 0 {
		env.configDirPath = path.Join(env.rootPath, "config")
	} else {
		env.configDirPath = path.Join(env.rootPath, opt.ConfigDirName)
	}

	return env
}
