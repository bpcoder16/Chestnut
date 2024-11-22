package cdefer

var deferFuncList []func()

func init() {
	deferFuncList = make([]func(), 0, 10)
}

func RegisterDeferFunc(df func()) {
	deferFuncList = append(deferFuncList, df)
}

func Defer() {
	for _, df := range deferFuncList {
		df()
	}
}
