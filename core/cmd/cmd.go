package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

func InitRootCmd(ctx context.Context) {
	// rootCmd 是主命令。
	rootCmd = &cobra.Command{
		Use:   "service-cli",
		Short: "A CLI with pluggable services",
		Long:  "This is a demo CLI application where services can be dynamically added or removed.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.SetContext(ctx)
		},
	}
}

// services 用于存储已注册的服务。
var services = make(map[string]Service)

func RegisterService(s Service) {
	services[s.Name()] = s
}

// generateServiceCommands 动态为所有注册的服务生成子命令。
func generateServiceCommands() {
	for _, service := range services {
		// 每个服务对应一个子命令。
		cmd := &cobra.Command{
			Use:   service.Name(),
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
