package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsrosen6/hyprdocked/internal/app"
	"github.com/dsrosen6/hyprdocked/internal/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	version = "0.3.0"
)

var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use: "hyprdocked",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			debug := viper.GetBool("debug")
			if debug {
				slog.SetLogLoggerLevel(slog.LevelDebug)
				slog.Debug("debug logging enabled")
			}
		},
	}

	versionCmd = &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}

	pingCmd = &cobra.Command{
		Use: "ping",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(app.SendPingCmd())
			fmt.Println("OK")
		},
	}

	idleCmd = &cobra.Command{
		Use:     "idle",
		Aliases: []string{"i"},
		Run: func(cmd *cobra.Command, args []string) {
			source, _ := cmd.Flags().GetString("source")
			cobra.CheckErr(app.SendIdleCmd(source))
		},
	}

	resumeCmd = &cobra.Command{
		Use:     "resume",
		Aliases: []string{"r"},
		Run: func(cmd *cobra.Command, args []string) {
			source, _ := cmd.Flags().GetString("source")
			cobra.CheckErr(app.SendResumeCmd(source))
		},
	}

	listenCmd = &cobra.Command{
		Use:     "listen",
		Aliases: []string{"l"},
		Run: func(cmd *cobra.Command, args []string) {
			var c app.Config
			cobra.CheckErr(viper.Unmarshal(&c))
			cobra.CheckErr(app.RunListener(c))
		},
	}

	serviceCmd = &cobra.Command{
		Use:   "service",
		Short: "Manage the hyprdocked systemd user service",
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

	checkCfgCmd = &cobra.Command{
		Use:   "check-cfg",
		Short: "Make sure config is valid and output values",
		Run: func(cmd *cobra.Command, args []string) {
			var cfg app.Config
			cobra.CheckErr(viper.Unmarshal(&cfg))

			var postHooks []string
			for _, h := range cfg.PostUpdateHooks {
				postHooks = append(postHooks, fmt.Sprintf("   On Status Change: %t, Command: %s", h.OnStatusChange, h.Command))
			}

			hookStr := "None"
			if len(postHooks) > 0 {
				hookStr = fmt.Sprintf("\n%s", strings.Join(postHooks, "\n"))
			}
			fmt.Printf("Debug: %v\nLaptop: %s\nSuspend On Idle: %v\nSuspend On Closed: %v\nPost Hooks: %s\n",
				cfg.Debug, cfg.Laptop, cfg.SuspendIdle, cfg.SuspendClosed, hookStr)
		},
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "enable debug logging")
	rootCmd.PersistentFlags().StringP("laptop", "l", "eDP-1", "laptop monitor name")
	rootCmd.PersistentFlags().Bool("suspend-idle", false, "suspend device when idle command is sent")
	rootCmd.PersistentFlags().Bool("suspend-closed", false, "suspend device on lid closed if only laptop")

	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("laptop", rootCmd.PersistentFlags().Lookup("laptop"))
	_ = viper.BindPFlag("suspend-idle", rootCmd.PersistentFlags().Lookup("suspend-idle"))
	_ = viper.BindPFlag("suspend-closed", rootCmd.PersistentFlags().Lookup("suspend-closed"))

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/hyprdocked/config.json)")
	idleCmd.Flags().String("source", "", "source of the idle command (logged by listener)")
	resumeCmd.Flags().String("source", "", "source of the resume command (logged by listener)")
	serviceInstallCmd.Flags().StringP("binary-path", "b", "", "custom binary path for the systemd unit to use")
	serviceLogsCmd.Flags().BoolP("stream", "f", false, "stream logs")

	serviceCmd.AddCommand(serviceInstallCmd)
	serviceCmd.AddCommand(serviceUninstallCmd)
	serviceCmd.AddCommand(serviceLogsCmd)
	serviceCmd.AddCommand(serviceRestartCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(idleCmd)
	rootCmd.AddCommand(resumeCmd)
	rootCmd.AddCommand(listenCmd)
	rootCmd.AddCommand(serviceCmd)
	rootCmd.AddCommand(checkCfgCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		cfgDir := filepath.Join(home, ".config")

		viper.AddConfigPath(filepath.Join(cfgDir, "hyprdocked"))
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			cobra.CheckErr(err)
		}
	}
}
