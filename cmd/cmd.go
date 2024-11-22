package cmd

import (
	"context"
	"fmt"
	"github.com/bpcoder16/Chestnut/appconfig/env"
	"github.com/bpcoder16/Chestnut/core/log"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

func InitRootCmd(ctx context.Context) {
	// rootCmd 是主命令。
	rootCmd = &cobra.Command{
		Use:   env.AppName() + "-Cli",
		Short: "命令应用列表",
		Long:  env.AppName() + " 的命令应用列表",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ctx = context.WithValue(ctx, log.DefaultMessageKey, "Command")
			ctx = context.WithValue(ctx, log.DefaultLogIdKey, uuid.New().String())
			cmd.SetContext(ctx)
		},
	}
}

// services 用于存储已注册的服务。
var services = make(map[string]Service)

func RegisterService(s Service) {
	services[s.Name(s)] = s
}

// generateServiceCommands 动态为所有注册的服务生成子命令。
func generateServiceCommands() {
	for _, service := range services {
		// 每个服务对应一个子命令。
		cmd := &cobra.Command{
			Use:   service.Name(service),
			Short: service.Description(),
			Run: func(svc Service) func(cmd *cobra.Command, args []string) {
				return func(cmd *cobra.Command, args []string) {
					svc.Run(cmd.Context(), args)
				}
			}(service),
		}
		rootCmd.AddCommand(cmd)
	}
}

func Run() {
	// 动态生成服务子命令
	generateServiceCommands()

	// 启动 CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("CMD Error:", err.Error())
	}
}
