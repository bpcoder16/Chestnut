package env

// Default (全局)默认的环境信息
//
// 全局的 RootDir() 、DataDir() 等方法均使用该环境信息
var Default = New(Option{})

// AppName (全局)应用的名称
func AppName() string {
	return Default.AppName()
}

// RunMode (全局) 程序运行等级
// 返回值 release 、test 、debug 之一
// 只能设置 'debug'、'test'、'release' 之一, 若是其他值，默认值会是 'debug'
func RunMode() string {
	return Default.RunMode()
}

func RootPath() string {
	return Default.RootPath()
}

func LocalIPV4() string {
	return Default.LocalIPV4()
}
