package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	PlanFile  string    `yaml:"plan_file" mapstructure:"plan_file"`
	Workspace string    `yaml:"workspace" mapstructure:"workspace"`
	Database  string    `yaml:"database" mapstructure:"database"`
	MCPPort   int       `yaml:"mcp_port" mapstructure:"mcp_port"`
	LLM       LLMConfig `yaml:"llm" mapstructure:"llm"`
	Agents    map[string]Agent `yaml:"agents" mapstructure:"agents"`
	Selection SelectionConfig `yaml:"selection" mapstructure:"selection"`
	Completion CompletionConfig `yaml:"completion" mapstructure:"completion"`
	Security  SecurityConfig `yaml:"security" mapstructure:"security"`
	Logging   LoggingConfig `yaml:"logging" mapstructure:"logging"`
	Development DevelopmentConfig `yaml:"development" mapstructure:"development"`
}

// LLMConfig represents LLM configuration
type LLMConfig struct {
	Primary        string      `yaml:"primary" mapstructure:"primary"`
	Fallback       *string     `yaml:"fallback" mapstructure:"fallback"`
	TimeoutSeconds int         `yaml:"timeout_seconds" mapstructure:"timeout_seconds"`
	MaxRetries     int         `yaml:"max_retries" mapstructure:"max_retries"`
	Claude         ClaudeConfig `yaml:"claude" mapstructure:"claude"`
	OpenAI         OpenAIConfig `yaml:"openai" mapstructure:"openai"`
}

// ClaudeConfig represents Claude Code configuration
type ClaudeConfig struct {
	Command       string   `yaml:"command" mapstructure:"command"`
	HeadlessArgs  []string `yaml:"headless_args" mapstructure:"headless_args"`
	OutputFormat  string   `yaml:"output_format" mapstructure:"output_format"`
	MCPConnect    bool     `yaml:"mcp_connect" mapstructure:"mcp_connect"`
}

// OpenAIConfig represents OpenAI CLI configuration
type OpenAIConfig struct {
	Command       string   `yaml:"command" mapstructure:"command"`
	HeadlessArgs  []string `yaml:"headless_args" mapstructure:"headless_args"`
}

// Agent represents an agent configuration
type Agent struct {
	Name          string            `yaml:"name" mapstructure:"name"`
	Role          string            `yaml:"role" mapstructure:"role"`
	AllowedStates []string          `yaml:"allowed_states" mapstructure:"allowed_states"`
	RoutingPolicy RoutingPolicy     `yaml:"routing_policy" mapstructure:"routing_policy"`
	Permissions   AgentPermissions  `yaml:"permissions" mapstructure:"permissions"`
}

// RoutingPolicy represents agent routing configuration
type RoutingPolicy struct {
	LLMPreference   string `yaml:"llm_preference" mapstructure:"llm_preference"`
	PromptTemplate  string `yaml:"prompt_template" mapstructure:"prompt_template"`
}

// AgentPermissions represents what an agent can do
type AgentPermissions struct {
	CanReadPlan         bool     `yaml:"can_read_plan" mapstructure:"can_read_plan"`
	CanExecuteCommands  bool     `yaml:"can_execute_commands" mapstructure:"can_execute_commands"`
	CanUpdateArtifacts  bool     `yaml:"can_update_artifacts" mapstructure:"can_update_artifacts"`
	CanReadArtifacts    bool     `yaml:"can_read_artifacts" mapstructure:"can_read_artifacts"`
	CanTransitionTo     []string `yaml:"can_transition_to" mapstructure:"can_transition_to"`
}

// SelectionConfig represents task selection policy
type SelectionConfig struct {
	Algorithm       string  `yaml:"algorithm" mapstructure:"algorithm"`
	PriorityWeight  float64 `yaml:"priority_weight" mapstructure:"priority_weight"`
	DependencyStrict bool   `yaml:"dependency_strict" mapstructure:"dependency_strict"`
	PreferLeafTasks bool    `yaml:"prefer_leaf_tasks" mapstructure:"prefer_leaf_tasks"`
	TieBreaker      string  `yaml:"tie_breaker" mapstructure:"tie_breaker"`
}

// CompletionConfig represents completion handshake settings
type CompletionConfig struct {
	MaxRetries                   int    `yaml:"max_retries" mapstructure:"max_retries"`
	RetryDelaySeconds           int    `yaml:"retry_delay_seconds" mapstructure:"retry_delay_seconds"`
	TimeoutSeconds              int    `yaml:"timeout_seconds" mapstructure:"timeout_seconds"`
	RequireExplicitStateUpdate  bool   `yaml:"require_explicit_state_update" mapstructure:"require_explicit_state_update"`
	FollowUpTemplate            string `yaml:"follow_up_template" mapstructure:"follow_up_template"`
}

// SecurityConfig represents security and safety settings
type SecurityConfig struct {
	AllowedCommands      []string `yaml:"allowed_commands" mapstructure:"allowed_commands"`
	WorkspaceRestriction bool     `yaml:"workspace_restriction" mapstructure:"workspace_restriction"`
	SecretPatterns       []string `yaml:"secret_patterns" mapstructure:"secret_patterns"`
	RedactInLogs         bool     `yaml:"redact_in_logs" mapstructure:"redact_in_logs"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level              string `yaml:"level" mapstructure:"level"`
	Format             string `yaml:"format" mapstructure:"format"`
	File               string `yaml:"file" mapstructure:"file"`
	AuditRetentionDays int    `yaml:"audit_retention_days" mapstructure:"audit_retention_days"`
}

// DevelopmentConfig represents development settings
type DevelopmentConfig struct {
	DryRunDefault         bool `yaml:"dry_run_default" mapstructure:"dry_run_default"`
	DebugMCP              bool `yaml:"debug_mcp" mapstructure:"debug_mcp"`
	CycleTimeboxSeconds   int  `yaml:"cycle_timebox_seconds" mapstructure:"cycle_timebox_seconds"`
}

// Load loads configuration from file and environment
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Set config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("baton")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.baton")
		v.AddConfigPath("/etc/baton")
	}

	// Environment variable support
	v.SetEnvPrefix("BATON")
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate and resolve paths
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// validate validates the configuration
func (c *Config) validate() error {
	// Resolve relative paths
	if !filepath.IsAbs(c.Database) {
		c.Database = filepath.Join(c.Workspace, c.Database)
	}

	if !filepath.IsAbs(c.PlanFile) {
		c.PlanFile = filepath.Join(c.Workspace, c.PlanFile)
	}

	// Validate workspace exists or can be created
	if err := os.MkdirAll(c.Workspace, 0755); err != nil {
		return fmt.Errorf("cannot create workspace directory %s: %w", c.Workspace, err)
	}

	// Validate port range
	if c.MCPPort < 1024 || c.MCPPort > 65535 {
		return fmt.Errorf("invalid MCP port %d: must be between 1024-65535", c.MCPPort)
	}

	return nil
}

// CreateDefaultConfig creates a default configuration file
func CreateDefaultConfig(path string) error {
	config := getDefaultConfig()

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	v.SetDefault("plan_file", "./plan.md")
	v.SetDefault("workspace", "./")
	v.SetDefault("database", "./baton.db")
	v.SetDefault("mcp_port", 8080)

	// LLM defaults
	v.SetDefault("llm.primary", "claude")
	v.SetDefault("llm.timeout_seconds", 300)
	v.SetDefault("llm.max_retries", 1)
	v.SetDefault("llm.claude.command", "claude")
	v.SetDefault("llm.claude.headless_args", []string{"-p"})
	v.SetDefault("llm.claude.output_format", "stream-json")
	v.SetDefault("llm.claude.mcp_connect", true)

	// Selection defaults
	v.SetDefault("selection.algorithm", "priority_dependency")
	v.SetDefault("selection.priority_weight", 1.0)
	v.SetDefault("selection.dependency_strict", true)
	v.SetDefault("selection.prefer_leaf_tasks", true)
	v.SetDefault("selection.tie_breaker", "oldest_updated")

	// Completion defaults
	v.SetDefault("completion.max_retries", 2)
	v.SetDefault("completion.retry_delay_seconds", 5)
	v.SetDefault("completion.timeout_seconds", 600)
	v.SetDefault("completion.require_explicit_state_update", true)
	v.SetDefault("completion.follow_up_template", "Are you finished? The state is not updated. Please either update the task state or provide a structured outcome with reason and next state.")

	// Security defaults
	v.SetDefault("security.allowed_commands", []string{"git", "npm", "go", "python", "pytest", "cargo", "make"})
	v.SetDefault("security.workspace_restriction", true)
	v.SetDefault("security.secret_patterns", []string{"sk-", "pk-", "token", "password", "secret"})
	v.SetDefault("security.redact_in_logs", true)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.file", "baton.log")
	v.SetDefault("logging.audit_retention_days", 90)

	// Development defaults
	v.SetDefault("development.dry_run_default", false)
	v.SetDefault("development.debug_mcp", false)
	v.SetDefault("development.cycle_timebox_seconds", 3600)
}