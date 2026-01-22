package config

import (
	"bytes"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
)

// Load reads YAML config file, then overrides via .env and environment variables.
// Env overrides are applied to the first item in each list for minimal usage.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	if err := ReadEnv(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	v := viper.New()
	v.SetConfigType("yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	data = []byte(os.ExpandEnv(string(data)))
	if err := v.ReadConfig(bytes.NewBuffer(data)); err != nil {
		return nil, err
	}
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	applyEnvOverrides(&cfg)
	applyGlobalOverrides(&cfg)
	expandDomainEnv(&cfg)

	return &cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if hasAnyEnv(checkEnvKeys()) {
		var ec CheckConfig
		if err := envconfig.Process("", &ec); err == nil {
			applyCheckOverrides(cfg, ec)
		}
	}
	if hasAnyEnv(policyEnvKeys()) {
		var pc PolicyConfig
		if err := envconfig.Process("", &pc); err == nil {
			applyPolicyOverrides(cfg, pc)
		}
	}
	if channelOverrideEnabled() {
		var cc ChannelConfig
		if err := envconfig.Process("", &cc); err == nil {
			applyChannelOverrides(cfg, cc)
		}
	}
	if hasAnyEnv(routeEnvKeys()) {
		var rc RouteMatch
		if err := envconfig.Process("", &rc); err == nil {
			applyRouteOverrides(cfg, rc)
		}
		if raw, ok := os.LookupEnv("ROUTE_TO"); ok {
			applyRouteToOverrides(cfg, raw)
		}
	}
	if hasAnyEnv(logEnvKeys()) {
		var lc LogConfig
		if err := envconfig.Process("", &lc); err == nil {
			applyLogOverrides(cfg, lc)
		}
	}
}

func applyCheckOverrides(cfg *Config, ec CheckConfig) {
	if len(cfg.Checks) == 0 {
		cfg.Checks = []CheckConfig{{}}
	}
	c := &cfg.Checks[0]

	if envNonEmpty("CHECK_TYPE") {
		c.Type = ec.Type
	}
	if envNonEmpty("CHECK_NAME") {
		c.Name = ec.Name
	}
	if envNonEmpty("CHECK_URL") {
		c.URL = ec.URL
	}
	if envNonEmpty("CHECK_INTERVAL") {
		c.Interval = ec.Interval
	}
	if envNonEmpty("CHECK_SCHEDULE") {
		c.Schedule = ec.Schedule
	}
	if envNonEmpty("CHECK_TIMEOUT") {
		c.Timeout = ec.Timeout
	}
	if envNonEmpty("CHECK_ADDRESS") {
		c.Address = ec.Address
	}
	if envNonEmpty("CHECK_SERVER_NAME") {
		c.ServerName = ec.ServerName
	}
	if envNonEmpty("CHECK_DOMAIN") {
		c.Domain = ec.Domain
	}
	if envNonEmpty("CHECK_TOKEN") {
		c.Token = ec.Token
	}
	if envNonEmpty("CHECK_RDAP_BASE_URL") {
		c.RDAPBaseURL = ec.RDAPBaseURL
	}
	if envNonEmpty("CHECK_RDAP_BASE_URLS") {
		c.RDAPBaseURLs = parseCSV(os.Getenv("CHECK_RDAP_BASE_URLS"))
	}
	if envNonEmpty("CHECK_WARN_BEFORE") {
		c.WarnBefore = ec.WarnBefore
	}
	if envNonEmpty("CHECK_CRIT_BEFORE") {
		c.CritBefore = ec.CritBefore
	}
	if envNonEmpty("CHECK_SKIP_VERIFY") {
		c.SkipVerify = ec.SkipVerify
	}
	if envNonEmpty("CHECK_NAMESPACE") {
		c.Namespace = ec.Namespace
	}
	if envNonEmpty("CHECK_LABEL_SELECTOR") {
		c.LabelSelector = ec.LabelSelector
	}
	if envNonEmpty("CHECK_KUBECONFIG") {
		c.Kubeconfig = ec.Kubeconfig
	}
	if envNonEmpty("CHECK_CONTEXT") {
		c.Context = ec.Context
	}
	if envNonEmpty("CHECK_MIN_READY") {
		c.MinReady = ec.MinReady
	}
}

func applyPolicyOverrides(cfg *Config, pc PolicyConfig) {
	if len(cfg.Policies) == 0 {
		cfg.Policies = []PolicyConfig{{}}
	}
	p := &cfg.Policies[0]

	if envNonEmpty("POLICY_NAME") {
		p.Name = pc.Name
	}
	if envNonEmpty("POLICY_COOLDOWN") {
		p.Cooldown = pc.Cooldown
	}
	if envNonEmpty("POLICY_NOTIFY_ON_RECOVERY") {
		p.NotifyOnRecovery = pc.NotifyOnRecovery
	}
}

func applyChannelOverrides(cfg *Config, cc ChannelConfig) {
	if len(cfg.Channels) == 0 {
		cfg.Channels = []ChannelConfig{{}}
	}
	ch := &cfg.Channels[0]

	typeSet := envNonEmpty("CHANNEL_TYPE")
	nameSet := envNonEmpty("CHANNEL_NAME")
	urlSet := envNonEmpty("CHANNEL_URL")
	timeoutSet := envNonEmpty("CHANNEL_TIMEOUT")
	usernameSet := envNonEmpty("CHANNEL_USERNAME")
	smtpHostSet := envNonEmpty("CHANNEL_SMTP_HOST")
	smtpPortSet := envNonEmpty("CHANNEL_SMTP_PORT")
	smtpUserSet := envNonEmpty("CHANNEL_SMTP_USERNAME")
	smtpPassSet := envNonEmpty("CHANNEL_SMTP_PASSWORD")
	smtpFromSet := envNonEmpty("CHANNEL_SMTP_FROM")
	smtpToSet := envNonEmpty("CHANNEL_SMTP_TO")
	smtpSubjectSet := envNonEmpty("CHANNEL_SMTP_SUBJECT")
	smtpImplicitSet := envNonEmpty("CHANNEL_SMTP_IMPLICIT_TLS")
	smtpSkipVerifySet := envNonEmpty("CHANNEL_SMTP_SKIP_VERIFY")

	if !typeSet && !nameSet && !urlSet && !timeoutSet && !usernameSet {
		if !smtpHostSet && !smtpPortSet && !smtpUserSet && !smtpPassSet && !smtpFromSet && !smtpToSet && !smtpSubjectSet && !smtpImplicitSet && !smtpSkipVerifySet {
			return
		}
	}

	if envNonEmpty("CHANNEL_TYPE") {
		ch.Type = cc.Type
	}
	if envNonEmpty("CHANNEL_NAME") {
		ch.Name = cc.Name
	}
	if envNonEmpty("CHANNEL_URL") {
		ch.URL = cc.URL
	}
	if envNonEmpty("CHANNEL_TIMEOUT") {
		ch.Timeout = cc.Timeout
	}
	if envNonEmpty("CHANNEL_USERNAME") {
		ch.Username = cc.Username
	}
	if smtpHostSet {
		ch.SMTPHost = cc.SMTPHost
	}
	if smtpPortSet {
		ch.SMTPPort = cc.SMTPPort
	}
	if smtpUserSet {
		ch.SMTPUsername = cc.SMTPUsername
	}
	if smtpPassSet {
		ch.SMTPPassword = cc.SMTPPassword
	}
	if smtpFromSet {
		ch.SMTPFrom = cc.SMTPFrom
	}
	if smtpToSet {
		ch.SMTPTo = cc.SMTPTo
	}
	if smtpSubjectSet {
		ch.SMTPSubject = cc.SMTPSubject
	}
	if smtpImplicitSet {
		ch.SMTPImplicitTLS = cc.SMTPImplicitTLS
	}
	if smtpSkipVerifySet {
		ch.SMTPSkipVerifyTLS = cc.SMTPSkipVerifyTLS
	}
}

func applyRouteOverrides(cfg *Config, rm RouteMatch) {
	if len(cfg.Routes) == 0 {
		cfg.Routes = []RouteConfig{{}}
	}
	r := &cfg.Routes[0]

	if envNonEmpty("ROUTE_MATCH_NAME") {
		r.Match.Name = rm.Name
	}
	if envNonEmpty("ROUTE_MATCH_STATUS") {
		r.Match.Status = rm.Status
	}
}

func applyRouteToOverrides(cfg *Config, raw string) {
	if len(cfg.Routes) == 0 {
		cfg.Routes = []RouteConfig{{}}
	}
	parts := strings.Split(raw, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) > 0 {
		cfg.Routes[0].To = out
	}
}

func hasAnyEnv(keys []string) bool {
	for _, key := range keys {
		if envNonEmpty(key) {
			return true
		}
	}
	return false
}

func checkEnvKeys() []string {
	return []string{
		"CHECK_TYPE", "CHECK_NAME", "CHECK_URL", "CHECK_INTERVAL", "CHECK_TIMEOUT",
		"CHECK_ADDRESS", "CHECK_SERVER_NAME", "CHECK_WARN_BEFORE", "CHECK_CRIT_BEFORE",
		"CHECK_DOMAIN", "CHECK_TOKEN", "CHECK_RDAP_BASE_URL", "CHECK_NAMESPACE", "CHECK_LABEL_SELECTOR",
		"CHECK_RDAP_BASE_URLS", "CHECK_KUBECONFIG", "CHECK_CONTEXT", "CHECK_MIN_READY", "CHECK_SCHEDULE",
		"CHECK_SKIP_VERIFY",
	}
}

func policyEnvKeys() []string {
	return []string{
		"POLICY_NAME", "POLICY_COOLDOWN", "POLICY_NOTIFY_ON_RECOVERY",
	}
}

func channelEnvKeys() []string {
	return []string{
		"CHANNEL_TYPE", "CHANNEL_NAME", "CHANNEL_URL", "CHANNEL_TIMEOUT", "CHANNEL_USERNAME",
		"CHANNEL_SMTP_HOST", "CHANNEL_SMTP_PORT", "CHANNEL_SMTP_USERNAME", "CHANNEL_SMTP_PASSWORD",
		"CHANNEL_SMTP_FROM", "CHANNEL_SMTP_TO", "CHANNEL_SMTP_SUBJECT",
		"CHANNEL_SMTP_IMPLICIT_TLS", "CHANNEL_SMTP_SKIP_VERIFY",
	}
}

func channelOverrideEnabled() bool {
	if envNonEmpty("CHANNEL_TYPE") || envNonEmpty("CHANNEL_NAME") || envNonEmpty("CHANNEL_URL") {
		return true
	}
	for _, key := range channelEnvKeys() {
		if envExists(key) {
			return true
		}
	}
	return false
}

func envExists(key string) bool {
	_, ok := os.LookupEnv(key)
	return ok
}

func routeEnvKeys() []string {
	return []string{
		"ROUTE_MATCH_NAME", "ROUTE_MATCH_STATUS", "ROUTE_TO",
	}
}

func logEnvKeys() []string {
	return []string{
		"LOG_LEVEL", "LOG_FORMAT", "LOG_FILE",
	}
}

func applyLogOverrides(cfg *Config, lc LogConfig) {
	if envNonEmpty("LOG_LEVEL") {
		cfg.Log.Level = lc.Level
	}
	if envNonEmpty("LOG_FORMAT") {
		cfg.Log.Format = lc.Format
	}
	if envNonEmpty("LOG_FILE") {
		cfg.Log.File = lc.File
	}
}

func applyGlobalOverrides(cfg *Config) {
	if envNonEmpty("PROBLEM_LIMIT") {
		if v, err := strconv.Atoi(strings.TrimSpace(os.Getenv("PROBLEM_LIMIT"))); err == nil {
			cfg.Notify.ProblemLimit = v
		}
	}
	if envNonEmpty("NOTIFY_AGGREGATE_BY_TYPE") {
		val := strings.TrimSpace(os.Getenv("NOTIFY_AGGREGATE_BY_TYPE"))
		if val == "1" || strings.EqualFold(val, "true") || strings.EqualFold(val, "yes") {
			cfg.Notify.AggregateByType = true
		}
	}
	if envNonEmpty("NOTIFY_AGGREGATE_WINDOW") {
		if d, err := time.ParseDuration(os.Getenv("NOTIFY_AGGREGATE_WINDOW")); err == nil {
			cfg.Notify.AggregateWindow = d
		}
	}
	if envNonEmpty("NOTIFY_STOP_ON_FAIL") {
		val := strings.TrimSpace(os.Getenv("NOTIFY_STOP_ON_FAIL"))
		if val == "1" || strings.EqualFold(val, "true") || strings.EqualFold(val, "yes") {
			cfg.Notify.StopOnFail = true
		}
	}
	if envNonEmpty("NOTIFY_RUN_ONCE") {
		val := strings.TrimSpace(os.Getenv("NOTIFY_RUN_ONCE"))
		if val == "1" || strings.EqualFold(val, "true") || strings.EqualFold(val, "yes") {
			cfg.Notify.RunOnce = true
		}
	}
}

func expandDomainEnv(cfg *Config) {
	raw, ok := os.LookupEnv("CHECK_DOMAINS")
	if !ok {
		return
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return
	}

	var warnBefore time.Duration
	var critBefore time.Duration
	var schedule string
	var timeout time.Duration
	var rdapBaseURL string
	var rdapBaseURLs []string

	if envNonEmpty("CHECK_DOMAIN_WARN_BEFORE") {
		if d, err := time.ParseDuration(os.Getenv("CHECK_DOMAIN_WARN_BEFORE")); err == nil {
			warnBefore = d
		}
	}
	if envNonEmpty("CHECK_DOMAIN_CRIT_BEFORE") {
		if d, err := time.ParseDuration(os.Getenv("CHECK_DOMAIN_CRIT_BEFORE")); err == nil {
			critBefore = d
		}
	}
	if envNonEmpty("CHECK_DOMAIN_SCHEDULE") {
		schedule = os.Getenv("CHECK_DOMAIN_SCHEDULE")
	}
	if envNonEmpty("CHECK_DOMAIN_TIMEOUT") {
		if d, err := time.ParseDuration(os.Getenv("CHECK_DOMAIN_TIMEOUT")); err == nil {
			timeout = d
		}
	}
	if envNonEmpty("CHECK_DOMAIN_RDAP_BASE_URL") {
		rdapBaseURL = os.Getenv("CHECK_DOMAIN_RDAP_BASE_URL")
	}
	if envNonEmpty("CHECK_DOMAIN_RDAP_BASE_URLS") {
		rdapBaseURLs = parseCSV(os.Getenv("CHECK_DOMAIN_RDAP_BASE_URLS"))
	}

	if schedule == "" {
		schedule = "0 3 * * *"
	}

	parts := strings.Split(raw, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		name := strings.ReplaceAll(p, ".", "-")
		cfg.Checks = append(cfg.Checks, CheckConfig{
			Type:         "domain_expiry",
			Name:         "domain-expiry-" + name,
			Domain:       p,
			WarnBefore:   warnBefore,
			CritBefore:   critBefore,
			Schedule:     schedule,
			Timeout:      timeout,
			RDAPBaseURL:  rdapBaseURL,
			RDAPBaseURLs: rdapBaseURLs,
		})
	}
}

func envNonEmpty(key string) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return false
	}
	return strings.TrimSpace(val) != ""
}

func parseCSV(input string) []string {
	parts := strings.Split(input, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
