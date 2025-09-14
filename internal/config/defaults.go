package config

// getDefaultConfig returns the complete default configuration
func getDefaultConfig() *Config {
	return &Config{
		PlanFile:  "./plan.md",
		Workspace: "./",
		Database:  "./baton.db",
		MCPPort:   8080,
		LLM: LLMConfig{
			Primary:        "claude",
			TimeoutSeconds: 300,
			MaxRetries:     1,
			Claude: ClaudeConfig{
				Command:      "claude",
				HeadlessArgs: []string{"-p"},
				OutputFormat: "stream-json",
				MCPConnect:   true,
			},
			OpenAI: OpenAIConfig{
				Command:      "openai",
				HeadlessArgs: []string{"--non-interactive"},
			},
		},
		Agents: map[string]Agent{
			"architect": {
				Name:          "System Architect",
				Role:          "Plans and designs system architecture",
				AllowedStates: []string{"ready_for_plan", "planning"},
				RoutingPolicy: RoutingPolicy{
					LLMPreference:  "claude",
					PromptTemplate: "architect.md",
				},
				Permissions: AgentPermissions{
					CanReadPlan:        true,
					CanUpdateArtifacts: true,
					CanTransitionTo:    []string{"planning", "ready_for_implementation"},
				},
			},
			"developer": {
				Name:          "Developer",
				Role:          "Implements code and fixes issues",
				AllowedStates: []string{"ready_for_implementation", "implementing", "fixing"},
				RoutingPolicy: RoutingPolicy{
					LLMPreference:  "claude",
					PromptTemplate: "developer.md",
				},
				Permissions: AgentPermissions{
					CanReadPlan:         true,
					CanExecuteCommands:  true,
					CanUpdateArtifacts:  true,
					CanTransitionTo:     []string{"implementing", "ready_for_code_review", "needs_fixes"},
				},
			},
			"reviewer": {
				Name:          "Code Reviewer",
				Role:          "Reviews code and provides feedback",
				AllowedStates: []string{"reviewing"},
				RoutingPolicy: RoutingPolicy{
					LLMPreference:  "claude",
					PromptTemplate: "reviewer.md",
				},
				Permissions: AgentPermissions{
					CanReadArtifacts:   true,
					CanUpdateArtifacts: true,
					CanTransitionTo:    []string{"ready_for_commit", "needs_fixes"},
				},
			},
		},
		Selection: SelectionConfig{
			Algorithm:        "priority_dependency",
			PriorityWeight:   1.0,
			DependencyStrict: true,
			PreferLeafTasks:  true,
			TieBreaker:       "oldest_updated",
		},
		Completion: CompletionConfig{
			MaxRetries:                  2,
			RetryDelaySeconds:          5,
			TimeoutSeconds:             600,
			RequireExplicitStateUpdate: true,
			FollowUpTemplate:           "Are you finished? The state is not updated. Please either update the task state or provide a structured outcome with reason and next state.",
		},
		Security: SecurityConfig{
			AllowedCommands:      []string{"git", "npm", "go", "python", "pytest", "cargo", "make"},
			WorkspaceRestriction: true,
			SecretPatterns:       []string{"sk-", "pk-", "token", "password", "secret"},
			RedactInLogs:         true,
		},
		Logging: LoggingConfig{
			Level:              "info",
			Format:             "json",
			File:               "baton.log",
			AuditRetentionDays: 90,
		},
		Development: DevelopmentConfig{
			DryRunDefault:       false,
			DebugMCP:            false,
			CycleTimeboxSeconds: 3600,
		},
	}
}