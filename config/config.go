package config

import (
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type (
	Config struct {
		Server    *Server    `mapstructure:"server" validate:"required"`
		Database  *Database  `mapstructure:"database" validate:"required"`
		Env       *Env       `mapstructure:"env" validate:"required"`
		Scheduler *Scheduler `mapstructure:"scheduler"`
	}

	Server struct {
		Port         int           `mapstructure:"port" validate:"required"`
		Environment  string        `mapstructure:"environment" validate:"required"`
		BodyLimit    int           `mapstructure:"bodyLimit" validate:"required"`
		TimeOut      time.Duration `mapstructure:"timeout" validate:"required"`
		ClientURL    string        `mapstructure:"client_url" validate:"required"`
		ServerDomain string        `mapstructure:"server_domain" validate:"required"`
	}

	Database struct {
		Driver   string `mapstructure:"driver" validate:"required"`
		Host     string `mapstructure:"host" validate:"required"`
		Port     int    `mapstructure:"port" validate:"required"`
		User     string `mapstructure:"user" validate:"required"`
		Password string `mapstructure:"password" validate:"required"`
		DBName   string `mapstructure:"dbname" validate:"required"`
		SSLMode  string `mapstructure:"sslmode" validate:"required"`
		Schema   string `mapstructure:"schema" validate:"required"`
	}

	Env struct {
		SecretKey        string `mapstructure:"secretkey" validate:"required"`
		OpenAIAPIKey     string `mapstructure:"openai_api_key"`
		AnthropicAPIKey  string `mapstructure:"anthropic_api_key"`
		GeminiAPIKey     string `mapstructure:"gemini_api_key"`
		PerplexityAPIKey string `mapstructure:"perplexity_api_key"`
		ClaudeAPIKey     string `mapstructure:"claude_api_key"`
		// RecommendationClaudeAPIKey is deliberately separate from ClaudeAPIKey
		// (the "claude" prompt-run engine) so Recommendations' Claude spend
		// shows up under its own key in the Anthropic console, independent of
		// the AI-agent engine's usage.
		RecommendationClaudeAPIKey string `mapstructure:"recommendation_claude_api_key"`
		// SerpApiAPIKey backs the "google-ai" engine — not an LLM call, it
		// queries Google Search via SerpApi and reads back whatever AI
		// Overview Google showed for that query.
		SerpApiAPIKey string `mapstructure:"serpapi_api_key"`
		ResendAPIKey  string `mapstructure:"resend_api_key"`
		// EmailFrom must be a Resend-verified sender; defaults to Resend's
		// shared test address so invites work before a custom domain is
		// verified.
		EmailFrom string `mapstructure:"email_from"`
	}

	Scheduler struct {
		Enabled        bool   `mapstructure:"enabled"`
		CronExpression string `mapstructure:"cron_expression"`
	}
)

var (
	once           sync.Once
	configInstance *Config
)

func GetConfig() *Config {
	once.Do(func() {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		// Allow env vars to override sensitive fields: DATABASE_PASSWORD, ENV_SECRETKEY, OPENAI_API_KEY, ANTHROPIC_API_KEY, GEMINI_API_KEY, PERPLEXITY_API_KEY
		viper.BindEnv("database.password", "DATABASE_PASSWORD")
		viper.BindEnv("env.secretkey", "ENV_SECRETKEY")
		viper.BindEnv("env.openai_api_key", "OPENAI_API_KEY")
		viper.BindEnv("env.anthropic_api_key", "ANTHROPIC_API_KEY")
		viper.BindEnv("env.gemini_api_key", "GEMINI_API_KEY")
		viper.BindEnv("env.perplexity_api_key", "PERPLEXITY_API_KEY")
		viper.BindEnv("env.claude_api_key", "CLAUDE_API_KEY")
		viper.BindEnv("env.recommendation_claude_api_key", "RECOMMENDATION_CLAUDE_API_KEY")
		viper.BindEnv("env.serpapi_api_key", "SERPAPI_API_KEY")
		viper.BindEnv("env.resend_api_key", "RESEND_API_KEY")
		viper.BindEnv("env.email_from", "EMAIL_FROM")
		viper.SetDefault("env.email_from", "Minimice Group Marketing AI <onboarding@resend.dev>")

		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}

		if err := viper.Unmarshal(&configInstance); err != nil {
			panic(err)
		}

		validating := validator.New()

		if err := validating.Struct(configInstance); err != nil {
			panic(err)
		}
	})

	return configInstance
}
