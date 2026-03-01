package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dsrosen6/hyprdocked/internal/app"
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
			err := app.SendPingCmd()
			cobra.CheckErr(err)

			fmt.Println("OK")
		},
	}

	idleCmd = &cobra.Command{
		Use:     "idle",
		Aliases: []string{"i"},
		Run: func(cmd *cobra.Command, args []string) {
			source, _ := cmd.Flags().GetString("source")
			err := app.SendIdleCmd(source)
			cobra.CheckErr(err)
		},
	}

	resumeCmd = &cobra.Command{
		Use:     "resume",
		Aliases: []string{"r"},
		Run: func(cmd *cobra.Command, args []string) {
			source, _ := cmd.Flags().GetString("source")
			err := app.SendResumeCmd(source)
			cobra.CheckErr(err)
		},
	}

	listenCmd = &cobra.Command{
		Use:     "listen",
		Aliases: []string{"l"},
		Run: func(cmd *cobra.Command, args []string) {
			p := app.ListenerParams{
				LaptopMonitorName: viper.GetString("laptop"),
				SuspendOnIdle:     viper.GetBool("suspend-idle"),
				SuspendOnClosed:   viper.GetBool("suspend-closed"),
			}

			err := app.RunListener(p)
			cobra.CheckErr(err)
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

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(idleCmd)
	rootCmd.AddCommand(resumeCmd)
	rootCmd.AddCommand(listenCmd)
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
		viper.SetConfigType("json")
		viper.SetConfigName("hyprdocked")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR reading config file:", err)
	}
}
