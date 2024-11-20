package cmd

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type Base struct {
}

func (b *Base) Name(child Service) string {
	t := reflect.TypeOf(child).Elem()
	return strings.ToLower(string(t.Name()[0])) + t.Name()[1:]
}

func (b *Base) Description() string {
	return "未设置描述内容（需要设置 Description() 方法）"
}

func (b *Base) Run(_ context.Context, _ []string) {
	fmt.Println("无具体实现功能")
}