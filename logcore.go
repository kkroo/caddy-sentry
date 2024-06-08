package sentry

import (
	"github.com/TheZeroSlave/zapsentry"
	"github.com/caddyserver/caddy/v2"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap/zapcore"
	"time"
)

func init() {
	caddy.RegisterModule(SentryLogCore{})
}

type SentryLogCore struct {
	Dsn                string            `json:"dsn,omitempty"`
	EnableTracing      bool              `json:"enable_tracing,omitempty"`
	TracesSampleRate   float64           `json:"traces_sample_rate,omitempty"`
	ProfilesSampleRate float64           `json:"profiles_sample_rate,omitempty"`
	Tags               map[string]string `json:"tags,omitempty"`

	sentryCore        zapcore.Core
	EnableBreadcrumbs bool
	client            *sentry.Client
}

// CaddyModule returns the Caddy module information.
func (SentryLogCore) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "caddy.logging.cores.sentry",
		New: func() caddy.Module { return new(SentryLogCore) },
	}
}

// Provision sets up the encoder.
func (fe *SentryLogCore) Provision(ctx caddy.Context) (err error) {
	clientOpts := sentry.ClientOptions{
		Dsn:                fe.Dsn,
		EnableTracing:      fe.EnableTracing,
		TracesSampleRate:   fe.TracesSampleRate,
		ProfilesSampleRate: fe.ProfilesSampleRate,
		Tags:               fe.Tags,
	}
	if err = sentry.Init(clientOpts); err != nil {
		return
	}
	if fe.client, err = sentry.NewClient(clientOpts); err != nil {
		return err
	}

	cfg := zapsentry.Configuration{
		Level:             zapcore.ErrorLevel,
		EnableBreadcrumbs: fe.EnableBreadcrumbs,
		BreadcrumbLevel:   zapcore.InfoLevel,
	}
	fe.sentryCore, err = zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromClient(fe.client))
	return
}

func (fe SentryLogCore) Cleanup() error {
	fe.client.Flush(2 * time.Second)
	return nil
}

func (fe SentryLogCore) Enabled(level zapcore.Level) bool {
	return fe.sentryCore.Enabled(level)
}

func (fe SentryLogCore) With(fields []zapcore.Field) zapcore.Core {
	return fe.sentryCore.With(fields)
}

func (fe SentryLogCore) Check(entry zapcore.Entry, entry2 *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return fe.sentryCore.Check(entry, entry2)
}

func (fe SentryLogCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	return fe.sentryCore.Write(entry, fields)
}

func (fe SentryLogCore) Sync() error {
	return fe.sentryCore.Sync()
}

// Interface guards
var (
	_ zapcore.Core       = (*SentryLogCore)(nil)
	_ caddy.Provisioner  = (*SentryLogCore)(nil)
	_ caddy.Module       = (*SentryLogCore)(nil)
	_ caddy.CleanerUpper = (*SentryLogCore)(nil)
)
