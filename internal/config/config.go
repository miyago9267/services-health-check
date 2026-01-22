package config

import "time"

type Config struct {
	Checks   []CheckConfig   `yaml:"checks" mapstructure:"checks"`
	Policies []PolicyConfig  `yaml:"policies" mapstructure:"policies"`
	Channels []ChannelConfig `yaml:"channels" mapstructure:"channels"`
	Routes   []RouteConfig   `yaml:"routes" mapstructure:"routes"`
	Log      LogConfig       `yaml:"log" mapstructure:"log"`
	Notify   NotifyConfig    `yaml:"notify" mapstructure:"notify"`
}

func DefaultConfig() Config {
	return Config{}
}

type CheckConfig struct {
	Type          string        `yaml:"type" mapstructure:"type" env:"CHECK_TYPE"`
	Name          string        `yaml:"name" mapstructure:"name" env:"CHECK_NAME"`
	URL           string        `yaml:"url" mapstructure:"url" env:"CHECK_URL"`
	Interval      time.Duration `yaml:"interval" mapstructure:"interval" env:"CHECK_INTERVAL"`
	Schedule      string        `yaml:"schedule" mapstructure:"schedule" env:"CHECK_SCHEDULE"`
	Timeout       time.Duration `yaml:"timeout" mapstructure:"timeout" env:"CHECK_TIMEOUT"`
	Address       string        `yaml:"address" mapstructure:"address" env:"CHECK_ADDRESS"`
	ServerName    string        `yaml:"server_name" mapstructure:"server_name" env:"CHECK_SERVER_NAME"`
	Domain        string        `yaml:"domain" mapstructure:"domain" env:"CHECK_DOMAIN"`
	Token         string        `yaml:"token" mapstructure:"token" env:"CHECK_TOKEN"`
	WarnBefore    time.Duration `yaml:"warn_before" mapstructure:"warn_before" env:"CHECK_WARN_BEFORE"`
	CritBefore    time.Duration `yaml:"crit_before" mapstructure:"crit_before" env:"CHECK_CRIT_BEFORE"`
	RDAPBaseURL   string        `yaml:"rdap_base_url" mapstructure:"rdap_base_url" env:"CHECK_RDAP_BASE_URL"`
	RDAPBaseURLs  []string      `yaml:"rdap_base_urls" mapstructure:"rdap_base_urls"`
	SkipVerify    bool          `yaml:"skip_verify" mapstructure:"skip_verify" env:"CHECK_SKIP_VERIFY"`
	Namespace     string        `yaml:"namespace" mapstructure:"namespace" env:"CHECK_NAMESPACE"`
	LabelSelector string        `yaml:"label_selector" mapstructure:"label_selector" env:"CHECK_LABEL_SELECTOR"`
	Kubeconfig    string        `yaml:"kubeconfig" mapstructure:"kubeconfig" env:"CHECK_KUBECONFIG"`
	Context       string        `yaml:"context" mapstructure:"context" env:"CHECK_CONTEXT"`
	MinReady      int           `yaml:"min_ready" mapstructure:"min_ready" env:"CHECK_MIN_READY"`
}

type PolicyConfig struct {
	Name             string        `yaml:"name" mapstructure:"name" env:"POLICY_NAME"`
	Cooldown         time.Duration `yaml:"cooldown" mapstructure:"cooldown" env:"POLICY_COOLDOWN"`
	NotifyOnRecovery bool          `yaml:"notify_on_recovery" mapstructure:"notify_on_recovery" env:"POLICY_NOTIFY_ON_RECOVERY"`
}

type ChannelConfig struct {
	Type              string        `yaml:"type" mapstructure:"type" env:"CHANNEL_TYPE"`
	Name              string        `yaml:"name" mapstructure:"name" env:"CHANNEL_NAME"`
	URL               string        `yaml:"url" mapstructure:"url" env:"CHANNEL_URL"`
	Timeout           time.Duration `yaml:"timeout" mapstructure:"timeout" env:"CHANNEL_TIMEOUT"`
	Username          string        `yaml:"username" mapstructure:"username" env:"CHANNEL_USERNAME"`
	SMTPHost          string        `yaml:"smtp_host" mapstructure:"smtp_host" env:"CHANNEL_SMTP_HOST"`
	SMTPPort          int           `yaml:"smtp_port" mapstructure:"smtp_port" env:"CHANNEL_SMTP_PORT"`
	SMTPUsername      string        `yaml:"smtp_username" mapstructure:"smtp_username" env:"CHANNEL_SMTP_USERNAME"`
	SMTPPassword      string        `yaml:"smtp_password" mapstructure:"smtp_password" env:"CHANNEL_SMTP_PASSWORD"`
	SMTPFrom          string        `yaml:"smtp_from" mapstructure:"smtp_from" env:"CHANNEL_SMTP_FROM"`
	SMTPTo            []string      `yaml:"smtp_to" mapstructure:"smtp_to" env:"CHANNEL_SMTP_TO"`
	SMTPSubject       string        `yaml:"smtp_subject" mapstructure:"smtp_subject" env:"CHANNEL_SMTP_SUBJECT"`
	SMTPImplicitTLS   bool          `yaml:"smtp_implicit_tls" mapstructure:"smtp_implicit_tls" env:"CHANNEL_SMTP_IMPLICIT_TLS"`
	SMTPSkipVerifyTLS bool          `yaml:"smtp_skip_verify" mapstructure:"smtp_skip_verify" env:"CHANNEL_SMTP_SKIP_VERIFY"`
}

type RouteConfig struct {
	Match RouteMatch `yaml:"match" mapstructure:"match"`
	To    []string   `yaml:"to" mapstructure:"to" env:"ROUTE_TO"`
}

type RouteMatch struct {
	Name   string `yaml:"name" mapstructure:"name" env:"ROUTE_MATCH_NAME"`
	Status string `yaml:"status" mapstructure:"status" env:"ROUTE_MATCH_STATUS"`
}

type NotifyConfig struct {
	ProblemLimit    int           `yaml:"problem_limit" mapstructure:"problem_limit" env:"PROBLEM_LIMIT"`
	AggregateByType bool          `yaml:"aggregate_by_type" mapstructure:"aggregate_by_type" env:"NOTIFY_AGGREGATE_BY_TYPE"`
	AggregateWindow time.Duration `yaml:"aggregate_window" mapstructure:"aggregate_window" env:"NOTIFY_AGGREGATE_WINDOW"`
	StopOnFail      bool          `yaml:"stop_on_fail" mapstructure:"stop_on_fail" env:"NOTIFY_STOP_ON_FAIL"`
	RunOnce         bool          `yaml:"run_once" mapstructure:"run_once" env:"NOTIFY_RUN_ONCE"`
}

type LogConfig struct {
	Level  string `yaml:"level" mapstructure:"level" env:"LOG_LEVEL"`
	Format string `yaml:"format" mapstructure:"format" env:"LOG_FORMAT"`
	File   string `yaml:"file" mapstructure:"file" env:"LOG_FILE"`
}
