package config

import "time"

type Config struct {
	Checks   []CheckConfig   `yaml:"checks"`
	Policies []PolicyConfig  `yaml:"policies"`
	Channels []ChannelConfig `yaml:"channels"`
	Routes   []RouteConfig   `yaml:"routes"`
	Log      LogConfig       `yaml:"log"`
}

func DefaultConfig() Config {
	return Config{}
}

type CheckConfig struct {
	Type          string        `yaml:"type" env:"CHECK_TYPE"`
	Name          string        `yaml:"name" env:"CHECK_NAME"`
	URL           string        `yaml:"url" env:"CHECK_URL"`
	Interval      time.Duration `yaml:"interval" env:"CHECK_INTERVAL"`
	Schedule      string        `yaml:"schedule" env:"CHECK_SCHEDULE"`
	Timeout       time.Duration `yaml:"timeout" env:"CHECK_TIMEOUT"`
	Address       string        `yaml:"address" env:"CHECK_ADDRESS"`
	ServerName    string        `yaml:"server_name" env:"CHECK_SERVER_NAME"`
	Domain        string        `yaml:"domain" env:"CHECK_DOMAIN"`
	Token         string        `yaml:"token" env:"CHECK_TOKEN"`
	WarnBefore    time.Duration `yaml:"warn_before" env:"CHECK_WARN_BEFORE"`
	CritBefore    time.Duration `yaml:"crit_before" env:"CHECK_CRIT_BEFORE"`
	SkipVerify    bool          `yaml:"skip_verify" env:"CHECK_SKIP_VERIFY"`
	Namespace     string        `yaml:"namespace" env:"CHECK_NAMESPACE"`
	LabelSelector string        `yaml:"label_selector" env:"CHECK_LABEL_SELECTOR"`
	Kubeconfig    string        `yaml:"kubeconfig" env:"CHECK_KUBECONFIG"`
	Context       string        `yaml:"context" env:"CHECK_CONTEXT"`
	MinReady      int           `yaml:"min_ready" env:"CHECK_MIN_READY"`
}

type PolicyConfig struct {
	Name             string        `yaml:"name" env:"POLICY_NAME"`
	Cooldown         time.Duration `yaml:"cooldown" env:"POLICY_COOLDOWN"`
	NotifyOnRecovery bool          `yaml:"notify_on_recovery" env:"POLICY_NOTIFY_ON_RECOVERY"`
}

type ChannelConfig struct {
	Type     string        `yaml:"type" env:"CHANNEL_TYPE"`
	Name     string        `yaml:"name" env:"CHANNEL_NAME"`
	URL      string        `yaml:"url" env:"CHANNEL_URL"`
	Timeout  time.Duration `yaml:"timeout" env:"CHANNEL_TIMEOUT"`
	Username string        `yaml:"username" env:"CHANNEL_USERNAME"`
}

type RouteConfig struct {
	Match RouteMatch `yaml:"match"`
	To    []string   `yaml:"to" env:"ROUTE_TO"`
}

type RouteMatch struct {
	Name   string `yaml:"name" env:"ROUTE_MATCH_NAME"`
	Status string `yaml:"status" env:"ROUTE_MATCH_STATUS"`
}

type LogConfig struct {
	Level  string `yaml:"level" env:"LOG_LEVEL"`
	Format string `yaml:"format" env:"LOG_FORMAT"`
	File   string `yaml:"file" env:"LOG_FILE"`
}
