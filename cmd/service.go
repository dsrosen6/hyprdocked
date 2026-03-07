package cmd

import (
	"github.com/dsrosen6/hyprdocked/internal/service"
	"github.com/spf13/cobra"
)

var (
	serviceCmd = &cobra.Command{
		Use:     "service",
		Aliases: []string{"svc"},
		Short:   "Manage the hyprdocked systemd user service",
	}

	serviceInstallCmd = &cobra.Command{
		Use:   "install",
		Short: "Install and start the hyprdocked systemd user service",
		Run: func(cmd *cobra.Command, args []string) {
			customBinary, _ := cmd.Flags().GetString("binary-path")
			cobra.CheckErr(service.Install(customBinary))
		},
	}

	serviceRestartCmd = &cobra.Command{
		Use:   "restart",
		Short: "Restart the hyprdocked systemd user service",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(service.Restart())
		},
	}

	serviceUninstallCmd = &cobra.Command{
		Use:   "uninstall",
		Short: "Stop, disable, and remove the hyprdocked systemd user service",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(service.Uninstall())
		},
	}

	serviceLogsCmd = &cobra.Command{
		Use:   "logs",
		Short: "Show logs of hyprdocked systemd user service",
		Run: func(cmd *cobra.Command, args []string) {
			stream, _ := cmd.Flags().GetBool("stream")
			cobra.CheckErr(service.ShowLogs(stream))
		},
	}
)

func init() {
	serviceInstallCmd.Flags().StringP("binary-path", "b", "", "custom binary path for the systemd unit to use")
	serviceLogsCmd.Flags().BoolP("stream", "f", false, "stream logs")

	serviceCmd.AddCommand(serviceInstallCmd)
	serviceCmd.AddCommand(serviceUninstallCmd)
	serviceCmd.AddCommand(serviceLogsCmd)
	serviceCmd.AddCommand(serviceRestartCmd)

	rootCmd.AddCommand(serviceCmd)
	rootCmd.AddCommand(serviceLogsCmd) // can also call as just hyprdocked logs
}
