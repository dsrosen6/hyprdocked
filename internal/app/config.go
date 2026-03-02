package app

type Config struct {
	Debug           bool       `mapstructure:"debug"`
	Laptop          string     `mapstructure:"laptop"`
	SuspendIdle     bool       `mapstructure:"suspend-idle"`
	SuspendClosed   bool       `mapstructure:"suspend-closed"`
	PostUpdateHooks []PostHook `mapstructure:"post-hooks"`
}

type PostHook struct {
	Command        string `mapstructure:"command"`
	OnStatusChange bool   `mapstructure:"on-status-change"`
}
