package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"services-health-check/internal/checkers/cloudflare"
	"services-health-check/internal/checkers/domain"
	httpcheck "services-health-check/internal/checkers/http"
	"services-health-check/internal/checkers/k8s"
	"services-health-check/internal/checkers/ssl"
	"services-health-check/internal/config"
	"services-health-check/internal/core/check"
	"services-health-check/internal/core/notify"
	"services-health-check/internal/core/policy"
	"services-health-check/internal/notifiers/discord"
	"services-health-check/internal/notifiers/gchat"
	"services-health-check/internal/notifiers/slack"
	"services-health-check/internal/notifiers/webhook"
	"services-health-check/internal/utils/logger"
)

type scheduledCheck struct {
	Checker    check.Checker
	Interval   time.Duration
	Schedule   string
	Type       string
	StopOnFail bool
	RunOnce    bool
}

func Run(ctx context.Context, configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log, closeLog, err := buildLogger(cfg.Log)
	if err != nil {
		return fmt.Errorf("logger: %w", err)
	}
	if closeLog != nil {
		defer closeLog()
	}
	log.Infof("config loaded: %s", configPath)

	checks, err := buildChecks(cfg)
	if err != nil {
		return fmt.Errorf("build checks: %w", err)
	}
	log.Infof("checks ready: %d", len(checks))

	for i, ch := range cfg.Channels {
		log.Infof("channel[%d]: name=%q type=%q", i, ch.Name, ch.Type)
	}

	notifiers, err := buildNotifiers(cfg)
	if err != nil {
		return fmt.Errorf("build notifiers: %w", err)
	}
	log.Infof("notifiers ready: %d", len(notifiers))

	pol := buildPolicy(cfg)

	results := make(chan check.Result)
	var wg sync.WaitGroup

	for _, sc := range checks {
		wg.Add(1)
		go func(sc scheduledCheck) {
			defer wg.Done()
			runCheckLoop(ctx, sc, results, log)
		}(sc)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var agg chan notify.Event
	if cfg.Notify.AggregateByType {
		agg = make(chan notify.Event, 100)
		expected := countChecksByType(cfg)
		go runAggregator(ctx, cfg, agg, notifiers, log, expected)
	}

	for res := range results {
		logResult(log, res)
		event, err := pol.Evaluate(ctx, res)
		if err != nil || event == nil {
			continue
		}
		event.Type = res.Type
		if agg != nil {
			agg <- *event
			continue
		}
		dispatch(ctx, cfg, *event, notifiers, log)
	}

	return nil
}

func buildLogger(cfg config.LogConfig) (*logger.Logger, func(), error) {
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.Format == "" {
		cfg.Format = "text"
	}

	if cfg.File == "" {
		return logger.New(logger.Config{Level: cfg.Level, Format: cfg.Format}), nil, nil
	}

	file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, err
	}

	closeFn := func() {
		_ = file.Close()
	}
	return logger.New(logger.Config{Level: cfg.Level, Format: cfg.Format, Output: file}), closeFn, nil
}

func buildChecks(cfg *config.Config) ([]scheduledCheck, error) {
	var checks []scheduledCheck
	for i, c := range cfg.Checks {
		switch c.Type {
		case "http":
			timeout := c.Timeout
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			checks = append(checks, scheduledCheck{
				Checker: &httpcheck.Checker{
					NameValue: c.Name,
					URL:       c.URL,
					Timeout:   timeout,
				},
				Interval:   c.Interval,
				Schedule:   c.Schedule,
				Type:       c.Type,
				StopOnFail: cfg.Notify.StopOnFail,
				RunOnce:    cfg.Notify.RunOnce,
			})
		case "k8s_pods":
			checks = append(checks, scheduledCheck{
				Checker: &k8s.PodChecker{
					NameValue:     c.Name,
					Namespace:     c.Namespace,
					LabelSelector: c.LabelSelector,
					Kubeconfig:    c.Kubeconfig,
					Context:       c.Context,
					MinReady:      c.MinReady,
					ProblemLimit:  cfg.Notify.ProblemLimit,
				},
				Interval:   c.Interval,
				Schedule:   c.Schedule,
				Type:       c.Type,
				StopOnFail: cfg.Notify.StopOnFail,
				RunOnce:    cfg.Notify.RunOnce,
			})
		case "ssl":
			timeout := c.Timeout
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			checks = append(checks, scheduledCheck{
				Checker: &ssl.Checker{
					NameValue:  c.Name,
					Address:    c.Address,
					ServerName: c.ServerName,
					Timeout:    timeout,
					WarnBefore: c.WarnBefore,
					CritBefore: c.CritBefore,
					SkipVerify: c.SkipVerify,
				},
				Interval:   c.Interval,
				Schedule:   c.Schedule,
				Type:       c.Type,
				StopOnFail: cfg.Notify.StopOnFail,
				RunOnce:    cfg.Notify.RunOnce,
			})
		case "cloudflare_token":
			timeout := c.Timeout
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			checks = append(checks, scheduledCheck{
				Checker: &cloudflare.TokenChecker{
					NameValue: c.Name,
					Token:     c.Token,
					Timeout:   timeout,
				},
				Interval:   c.Interval,
				Schedule:   c.Schedule,
				Type:       c.Type,
				StopOnFail: cfg.Notify.StopOnFail,
				RunOnce:    cfg.Notify.RunOnce,
			})
		case "domain_expiry":
			timeout := c.Timeout
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			checks = append(checks, scheduledCheck{
				Checker: &domain.ExpiryChecker{
					NameValue:   c.Name,
					Domain:      c.Domain,
					Timeout:     timeout,
					WarnBefore:  c.WarnBefore,
					CritBefore:  c.CritBefore,
					RDAPBaseURL: c.RDAPBaseURL,
				},
				Interval:   c.Interval,
				Schedule:   c.Schedule,
				Type:       c.Type,
				StopOnFail: cfg.Notify.StopOnFail,
				RunOnce:    cfg.Notify.RunOnce,
			})
		default:
			return nil, fmt.Errorf("unknown check type at index %d (name=%q): %q", i, c.Name, c.Type)
		}
	}
	return checks, nil
}

func buildNotifiers(cfg *config.Config) (map[string]notify.Notifier, error) {
	notifiers := make(map[string]notify.Notifier)
	for i, c := range cfg.Channels {
		switch c.Type {
		case "webhook":
			timeout := c.Timeout
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			notifiers[c.Name] = &webhook.Notifier{
				NameValue: c.Name,
				URL:       c.URL,
				Timeout:   timeout,
			}
		case "discord":
			timeout := c.Timeout
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			notifiers[c.Name] = &discord.Notifier{
				NameValue: c.Name,
				URL:       c.URL,
				Username:  c.Username,
				Timeout:   timeout,
			}
		case "slack":
			timeout := c.Timeout
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			notifiers[c.Name] = &slack.Notifier{
				NameValue: c.Name,
				URL:       c.URL,
				Timeout:   timeout,
			}
		case "gchat":
			timeout := c.Timeout
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			notifiers[c.Name] = &gchat.Notifier{
				NameValue: c.Name,
				URL:       c.URL,
				Timeout:   timeout,
			}
		default:
			return nil, fmt.Errorf("unknown channel type at index %d (name=%q): %q", i, c.Name, c.Type)
		}
	}
	return notifiers, nil
}

func buildPolicy(cfg *config.Config) *policy.SimplePolicy {
	var polCfg config.PolicyConfig
	if len(cfg.Policies) > 0 {
		polCfg = cfg.Policies[0]
	}
	return policy.NewSimplePolicy(polCfg.Cooldown, polCfg.NotifyOnRecovery)
}

func runCheckLoop(ctx context.Context, sc scheduledCheck, results chan<- check.Result, log *logger.Logger) {
	status := runOnce(ctx, sc.Checker, results, sc.Type)
	if sc.RunOnce {
		return
	}
	if sc.StopOnFail && status != check.StatusOK {
		return
	}

	if sc.Schedule != "" {
		runCronLoop(ctx, sc, results, log)
		return
	}

	interval := sc.Interval
	if interval == 0 {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			status = runOnce(ctx, sc.Checker, results, sc.Type)
			if sc.StopOnFail && status != check.StatusOK {
				return
			}
		}
	}
}

func runCronLoop(ctx context.Context, sc scheduledCheck, results chan<- check.Result, log *logger.Logger) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	c := cron.New(cron.WithParser(parser))
	_, err := c.AddFunc(sc.Schedule, func() {
		status := runOnce(ctx, sc.Checker, results, sc.Type)
		if sc.StopOnFail && status != check.StatusOK {
			return
		}
	})
	if err != nil {
		log.Errorf("invalid schedule for %q: %v", sc.Checker.Name(), err)
		return
	}
	c.Start()
	defer c.Stop()

	<-ctx.Done()
}

func runOnce(ctx context.Context, checker check.Checker, results chan<- check.Result, checkType string) check.Status {
	if ctx.Err() != nil {
		return check.StatusUnknown
	}
	res, err := checker.Check(ctx)
	res.Type = checkType
	if err != nil {
		if ctx.Err() == nil {
			results <- res
		}
		return res.Status
	}
	if ctx.Err() == nil {
		results <- res
	}
	return res.Status
}

func runAggregator(ctx context.Context, cfg *config.Config, in <-chan notify.Event, notifiers map[string]notify.Notifier, log *logger.Logger, expected map[string]int) {
	window := cfg.Notify.AggregateWindow
	if window == 0 {
		window = 30 * time.Second
	}
	ticker := time.NewTicker(window)
	defer ticker.Stop()

	buffer := make(map[string][]notify.Event)
	flushAll := func() {
		for key, items := range buffer {
			if len(items) == 0 {
				continue
			}
			aggregateAndDispatch(ctx, cfg, key, items, notifiers, log)
		}
		buffer = make(map[string][]notify.Event)
	}
	flushType := func(key string) {
		items := buffer[key]
		if len(items) == 0 {
			return
		}
		aggregateAndDispatch(ctx, cfg, key, items, notifiers, log)
		delete(buffer, key)
	}

	for {
		select {
		case <-ctx.Done():
			flushAll()
			return
		case ev := <-in:
			key := ev.Type
			if key == "" {
				key = "unknown"
			}
			buffer[key] = append(buffer[key], ev)
			if expected[key] > 0 && len(buffer[key]) >= expected[key] {
				flushType(key)
			}
		case <-ticker.C:
			flushAll()
		}
	}
}

func aggregateAndDispatch(ctx context.Context, cfg *config.Config, key string, items []notify.Event, notifiers map[string]notify.Notifier, log *logger.Logger) {
	status := highestStatus(items)
	summary := fmt.Sprintf("%s 檢查彙總（%d）", typeLabel(key), len(items))
	details := buildAggregateDetails(items)
	agg := notify.Event{
		Service:    key,
		Type:       key,
		Status:     status,
		Summary:    summary,
		Details:    details,
		OccurredAt: time.Now(),
	}
	dispatch(ctx, cfg, agg, notifiers, log)
}

func highestStatus(events []notify.Event) string {
	priority := map[string]int{"CRIT": 3, "WARN": 2, "OK": 1, "UNKNOWN": 0}
	best := "UNKNOWN"
	bestScore := -1
	for _, ev := range events {
		score := priority[ev.Status]
		if score > bestScore {
			bestScore = score
			best = ev.Status
		}
	}
	return best
}

func buildAggregateDetails(events []notify.Event) string {
	var lines []string
	for _, ev := range events {
		if ev.Status == "OK" {
			continue
		}
		detail := ev.Details
		if ev.Type != "domain_expiry" {
			detail = fmt.Sprintf("%s: %s", ev.Service, ev.Details)
		}
		line := fmt.Sprintf("[%s] %s", ev.Status, detail)
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return "無 WARN/CRIT"
	}
	return strings.Join(lines, "; ")
}

func typeLabel(key string) string {
	switch key {
	case "k8s_pods":
		return "K8s Pod"
	case "ssl":
		return "SSL"
	case "domain_expiry":
		return "Domain"
	case "cloudflare_token":
		return "Cloudflare Token"
	case "http":
		return "HTTP"
	default:
		return key
	}
}

func countChecksByType(cfg *config.Config) map[string]int {
	out := make(map[string]int)
	for _, c := range cfg.Checks {
		if c.Type == "" {
			continue
		}
		out[c.Type]++
	}
	return out
}

func dispatch(ctx context.Context, cfg *config.Config, event notify.Event, notifiers map[string]notify.Notifier, log *logger.Logger) {
	for _, route := range cfg.Routes {
		if !matchRoute(route.Match, event) {
			continue
		}
		for _, name := range route.To {
			n, ok := notifiers[name]
			if !ok {
				continue
			}
			if ctx.Err() != nil {
				return
			}
			if err := n.Send(ctx, event); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				log.Errorf("notify %s: %v", name, err)
				continue
			}
			log.Infof("notify %s: %s %s", name, event.Service, event.Status)
		}
	}
}

func matchRoute(match config.RouteMatch, event notify.Event) bool {
	if match.Name != "" && match.Name != event.Service {
		return false
	}
	if match.Status != "" && match.Status != event.Status {
		return false
	}
	return true
}

func logResult(log *logger.Logger, res check.Result) {
	switch res.Status {
	case check.StatusCrit:
		log.Errorf("結果 %s: %s", res.Name, res.Message)
	case check.StatusWarn:
		log.Warnf("結果 %s: %s", res.Name, res.Message)
	case check.StatusOK:
		log.Infof("結果 %s: %s", res.Name, res.Message)
	default:
		log.Infof("結果 %s: %s", res.Name, res.Message)
	}
}
