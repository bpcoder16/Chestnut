package env

import "time"

var Default = New(Option{})

func RootPath() string {
	return Default.RootPath()
}

func LocalIPV4() string {
	return Default.LocalIPV4()
}

func ConfigMode() string {
	return Default.ConfigMode()
}

func AppName() string {
	return Default.AppName()
}

func RunMode() string {
	return Default.RunMode()
}

func TimeLocation() *time.Location {
	return Default.TimeLocation()
}
