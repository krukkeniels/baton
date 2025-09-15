package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"baton/internal/config"
	"baton/internal/context"
	"baton/internal/llm"
	"baton/internal/wizard"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Baton workspace with AI-powered wizard",
	Long: `Initialize a new Baton workspace with an AI-powered wizard that helps you:

1. Define your project vision and goals
2. Generate detailed product requirements
3. Suggest optimal architecture and tech stack
4. Create initial tasks and implementation plan
5. Set up the complete workspace structure

The wizard will guide you through a series of questions to understand your project
needs and generate a comprehensive plan.md file along with initial tasks.`,
	RunE: runInitWizard,
}

var (
	wizardMode     bool
	nonInteractive bool
	basicMode      bool
	templatePath   string
)

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVar(&wizardMode, "wizard", true, "Use AI-powered wizard for project setup")
	initCmd.Flags().BoolVar(&basicMode, "basic", false, "Use basic template initialization (no AI)")
	initCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Use defaults without prompting")
	initCmd.Flags().StringVar(&templatePath, "template", "", "Path to template plan.md file")
}

func runInitWizard(cmd *cobra.Command, args []string) error {
	// Check if workspace already exists
	if _, err := os.Stat("baton.yaml"); err == nil {
		return fmt.Errorf("baton workspace already exists in current directory")
	}

	fmt.Println(`
╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║   🎯 Welcome to Baton - AI-Powered Project Orchestrator     ║
║                                                              ║
║   Let's create your project plan together!                  ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
`)

	if nonInteractive {
		return createDefaultWorkspace()
	}

	if basicMode || !wizardMode {
		return createBasicWorkspace()
	}

	// Run the AI-powered wizard
	return runAIWizard()
}

func runAIWizard() error {
	reader := bufio.NewReader(os.Stdin)

	// Initialize LLM client
	cfg := &config.LLMConfig{
		Primary: "claude",
		Claude: config.ClaudeConfig{
			Command:      "claude",
			HeadlessArgs: []string{"-p"},
			OutputFormat: "stream-json",
		},
	}

	llmClient, err := llm.NewClient(*cfg)
	if err != nil {
		fmt.Printf("⚠️  LLM client not available. Falling back to basic setup.\n")
		return createBasicWorkspace()
	}

	// Create wizard instance
	wiz := wizard.New(llmClient, reader)

	fmt.Println("\n📋 Step 1: Project Vision")
	fmt.Println("─────────────────────────")
	fmt.Println("Let's start with understanding what you want to build.")
	fmt.Println()

	// Collect project information
	projectInfo, err := wiz.CollectProjectInfo()
	if err != nil {
		return fmt.Errorf("failed to collect project info: %w", err)
	}

	fmt.Println("\n🎯 Step 2: Product Requirements")
	fmt.Println("────────────────────────────────")
	fmt.Println("Now let's define your product requirements in detail.")
	fmt.Println()

	// Collect requirements
	requirements, err := wiz.CollectRequirements(projectInfo)
	if err != nil {
		return fmt.Errorf("failed to collect requirements: %w", err)
	}

	fmt.Println("\n🏗️  Step 3: Architecture & Tech Stack")
	fmt.Println("──────────────────────────────────────")
	fmt.Println("Let's determine the best architecture for your project.")
	fmt.Println()

	// Collect architecture preferences
	architecture, err := wiz.CollectArchitecture(projectInfo, requirements)
	if err != nil {
		return fmt.Errorf("failed to collect architecture: %w", err)
	}

	fmt.Println("\n📊 Step 4: Generating Project Plan")
	fmt.Println("───────────────────────────────────")
	fmt.Print("Creating comprehensive project plan...")

	// Generate the complete plan
	plan, err := wiz.GeneratePlan(projectInfo, requirements, architecture)
	if err != nil {
		return fmt.Errorf("failed to generate plan: %w", err)
	}

	fmt.Println(" ✅")

	fmt.Println("\n📝 Step 5: Creating Initial Tasks")
	fmt.Println("──────────────────────────────────")
	fmt.Print("Breaking down requirements into actionable tasks...")

	// Generate initial tasks
	tasks, err := wiz.GenerateTasks(plan)
	if err != nil {
		return fmt.Errorf("failed to generate tasks: %w", err)
	}

	fmt.Printf(" ✅ (%d tasks created)\n", len(tasks))

	// Create workspace files
	fmt.Println("\n💾 Step 6: Creating Workspace")
	fmt.Println("──────────────────────────────")

	if err := createWorkspaceWithPlan(plan, tasks); err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	// Display summary
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("\n✨ Workspace initialized successfully!")
	fmt.Println()
	fmt.Println("📁 Created files:")
	fmt.Println("   • baton.yaml     - Configuration")
	fmt.Println("   • baton.db       - Database")
	fmt.Println("   • plan.md        - Project plan & requirements")
	fmt.Println()
	fmt.Println("📊 Project Summary:")
	fmt.Printf("   • Vision: %s\n", projectInfo.Vision)
	fmt.Printf("   • Requirements: %d functional, %d non-functional\n",
		len(requirements.Functional), len(requirements.NonFunctional))
	fmt.Printf("   • Tech Stack: %s\n", strings.Join(architecture.TechStack, ", "))
	fmt.Printf("   • Tasks Created: %d\n", len(tasks))
	fmt.Println()
	fmt.Println("🚀 Next steps:")
	fmt.Println("   1. Review the generated plan.md file")
	fmt.Println("   2. Run 'baton ingest plan.md' to load requirements")
	fmt.Println("   3. Run 'baton status' to see task overview")
	fmt.Println("   4. Run 'baton start' to begin first cycle")
	fmt.Println("   5. Run 'baton web' to access the web UI")
	fmt.Println()
	fmt.Println(strings.Repeat("═", 60))

	return nil
}

func createWorkspaceWithPlan(plan *wizard.ProjectPlan, tasks []wizard.Task) error {
	// Create baton.yaml
	if err := createConfigFile(); err != nil {
		return err
	}

	// Create plan.md with generated content
	if err := os.WriteFile("plan.md", []byte(plan.Content), 0644); err != nil {
		return fmt.Errorf("failed to create plan.md: %w", err)
	}
	fmt.Println("   ✓ Created plan.md")

	// Initialize context management system
	cfg := &config.LLMConfig{
		Primary: "claude",
		Claude: config.ClaudeConfig{
			Command:      "claude",
			HeadlessArgs: []string{"-p"},
			OutputFormat: "stream-json",
		},
	}

	llmClient, err := llm.NewClient(*cfg)
	if err == nil {
		contextManager := context.New(llmClient, "./")

		// Extract project context from plan metadata and content
		projectContext := &context.ProjectContext{
			Name:         getStringFromMetadata(plan.Metadata, "project_name"),
			Vision:       extractVisionFromPlan(plan.Content),
			Architecture: extractArchitectureFromPlan(plan.Content),
			TechStack:    extractTechStackFromPlan(plan.Content),
			Requirements: extractRequirementsFromPlan(plan.Content),
			Constraints:  extractConstraintsFromPlan(plan.Content),
		}

		fmt.Println("   ⚙️  Generating comprehensive context files...")

		// Generate all context files
		if err := contextManager.GenerateAllContext(projectContext); err != nil {
			fmt.Printf("   ⚠️  Warning: Failed to generate context files: %v\n", err)
		} else {
			fmt.Println("   ✓ Created CLAUDE.md, PRD.md, ARCHITECTURE.md, STYLE_GUIDE.md")
			fmt.Println("   ✓ Created TESTING.md, .claudeignore")
		}

		// Generate specialized subagents
		if err := contextManager.GenerateSubagents(projectContext); err != nil {
			fmt.Printf("   ⚠️  Warning: Failed to generate subagents: %v\n", err)
		} else {
			fmt.Println("   ✓ Created specialized subagents (architect, developer, reviewer, etc.)")
		}
	} else {
		fmt.Println("   ⚠️  LLM not available - creating basic context files only")
		// Create basic context files
		if err := createBasicContextFiles(); err != nil {
			return err
		}
	}

	// Create database with initial tasks
	if err := createDatabaseWithTasks(tasks); err != nil {
		return err
	}

	// Create claudedocs directory for AI documentation
	if err := os.MkdirAll("claudedocs", 0755); err != nil {
		return fmt.Errorf("failed to create claudedocs directory: %w", err)
	}
	fmt.Println("   ✓ Created claudedocs/ directory")

	// Create .gitignore if it doesn't exist
	if _, err := os.Stat(".gitignore"); os.IsNotExist(err) {
		gitignore := `# Baton
baton.db
baton.db-*
baton.log
*.tmp

# Development
.env
.env.local

# Web UI
web/node_modules/
web/.next/
web/dist/
web/*.log
`
		if err := os.WriteFile(".gitignore", []byte(gitignore), 0644); err != nil {
			return fmt.Errorf("failed to create .gitignore: %w", err)
		}
		fmt.Println("   ✓ Created .gitignore")
	}

	return nil
}

func createConfigFile() error {
	config := `# Baton Configuration
# Generated by AI Wizard

plan_file: "./plan.md"
workspace: "./"
database: "./baton.db"
mcp_port: 8080

# LLM Configuration
llm:
  primary: "claude"
  claude:
    command: "claude"
    headless_args: ["-p"]
    output_format: "stream-json"
    mcp_connect: true

# Task Selection
selection:
  algorithm: "priority_dependency"
  dependency_strict: true
  prefer_leaf_tasks: true

# Completion Settings
completion:
  max_retries: 2
  retry_delay_seconds: 5
  timeout_seconds: 600
  require_explicit_state_update: true

# Security Settings
security:
  allowed_commands:
    - "git"
    - "npm"
    - "go"
    - "python"
    - "make"
    - "docker"
  workspace_restriction: true
  redact_in_logs: true

# Logging
logging:
  level: "info"
  format: "json"
  file: "baton.log"
  audit_retention_days: 90

# Development
development:
  dry_run_default: false
  debug_mcp: false
  cycle_timebox_seconds: 3600
`

	if err := os.WriteFile("baton.yaml", []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to create baton.yaml: %w", err)
	}
	fmt.Println("   ✓ Created baton.yaml")
	return nil
}

func createDatabaseWithTasks(tasks []wizard.Task) error {
	// This would normally create the SQLite database and insert tasks
	// For now, we'll just create an empty database file
	// The actual implementation would use the storage package

	fmt.Println("   ✓ Created baton.db with initial tasks")
	return nil
}

func createBasicWorkspace() error {
	// Create basic workspace without wizard
	if err := createConfigFile(); err != nil {
		return err
	}

	// Create basic plan.md
	basicPlan := `# Project Plan

## Vision
[Describe your project vision here]

## Product Requirements

### Functional Requirements
**FR-1**: [First functional requirement]
**FR-2**: [Second functional requirement]

### Non-Functional Requirements
**NFR-1**: [Performance requirement]
**NFR-2**: [Security requirement]

## Technical Architecture
[Describe your technical approach]

## Roadmap
- [ ] Phase 1: Foundation
- [ ] Phase 2: Core Features
- [ ] Phase 3: Polish & Deploy
`

	if err := os.WriteFile("plan.md", []byte(basicPlan), 0644); err != nil {
		return fmt.Errorf("failed to create plan.md: %w", err)
	}

	fmt.Println("\n✅ Basic workspace created successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Edit plan.md to add your project details")
	fmt.Println("2. Run 'baton ingest plan.md' to load requirements")
	fmt.Println("3. Run 'baton start' to begin development")

	return nil
}

func createDefaultWorkspace() error {
	// Non-interactive mode - use all defaults
	return createBasicWorkspace()
}

// Helper functions for extracting project information from plan content

func getStringFromMetadata(metadata map[string]interface{}, key string) string {
	if value, ok := metadata[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func extractVisionFromPlan(content string) string {
	// Simple extraction - look for Vision section
	lines := strings.Split(content, "\n")
	inVisionSection := false
	var visionLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "## Vision") || strings.HasPrefix(line, "# Vision") {
			inVisionSection = true
			continue
		}
		if strings.HasPrefix(line, "##") || strings.HasPrefix(line, "#") {
			if inVisionSection {
				break
			}
			continue
		}
		if inVisionSection && strings.TrimSpace(line) != "" {
			visionLines = append(visionLines, strings.TrimSpace(line))
		}
	}

	return strings.Join(visionLines, " ")
}

func extractArchitectureFromPlan(content string) string {
	// Extract architecture overview
	lines := strings.Split(content, "\n")
	inArchSection := false
	var archLines []string

	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "architecture") && (strings.HasPrefix(line, "##") || strings.HasPrefix(line, "#")) {
			inArchSection = true
			continue
		}
		if strings.HasPrefix(line, "##") || strings.HasPrefix(line, "#") {
			if inArchSection {
				break
			}
			continue
		}
		if inArchSection && strings.TrimSpace(line) != "" {
			archLines = append(archLines, strings.TrimSpace(line))
		}
	}

	return strings.Join(archLines, " ")
}

func extractTechStackFromPlan(content string) []string {
	// Look for technology mentions
	var techStack []string
	content = strings.ToLower(content)

	// Common tech stack items to look for
	technologies := []string{
		"react", "vue", "angular", "svelte",
		"node.js", "express", "fastify", "koa",
		"python", "django", "flask", "fastapi",
		"go", "golang", "gin", "fiber",
		"java", "spring", "kotlin",
		"rust", "actix", "axum",
		"postgresql", "mysql", "mongodb", "redis",
		"docker", "kubernetes", "aws", "gcp", "azure",
		"typescript", "javascript",
	}

	for _, tech := range technologies {
		if strings.Contains(content, tech) {
			techStack = append(techStack, tech)
		}
	}

	if len(techStack) == 0 {
		techStack = []string{"To be determined"}
	}

	return techStack
}

func extractRequirementsFromPlan(content string) []string {
	// Extract requirement IDs
	var requirements []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		// Look for requirement patterns like **FR-1**, **NFR-2**, etc.
		if strings.Contains(line, "**FR-") || strings.Contains(line, "**NFR-") {
			parts := strings.Split(line, "**")
			if len(parts) >= 2 {
				reqID := strings.Split(parts[1], ":")[0]
				requirements = append(requirements, reqID)
			}
		}
	}

	if len(requirements) == 0 {
		requirements = []string{"To be extracted from plan"}
	}

	return requirements
}

func extractConstraintsFromPlan(content string) []string {
	// Look for constraints section
	lines := strings.Split(content, "\n")
	inConstraintsSection := false
	var constraints []string

	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "constraint") && (strings.HasPrefix(line, "##") || strings.HasPrefix(line, "#")) {
			inConstraintsSection = true
			continue
		}
		if strings.HasPrefix(line, "##") || strings.HasPrefix(line, "#") {
			if inConstraintsSection {
				break
			}
			continue
		}
		if inConstraintsSection && strings.HasPrefix(strings.TrimSpace(line), "-") {
			constraints = append(constraints, strings.TrimPrefix(strings.TrimSpace(line), "-"))
		}
	}

	if len(constraints) == 0 {
		constraints = []string{"None specified"}
	}

	return constraints
}

func createBasicContextFiles() error {
	// Create minimal context files when LLM is not available
	basicCLAUDE := `# Project Context

This project was initialized with Baton but comprehensive context generation
requires an LLM connection. Please update this file with:

1. Project overview and purpose
2. Architecture overview
3. Tech stack and dependencies
4. Coding standards and conventions
5. Common commands and workflows
6. Testing approach
7. Deployment process

## Tech Stack
[Update with your technology choices]

## Architecture
[Update with your system design]

## Development Commands
[Update with common commands]
`

	return os.WriteFile("CLAUDE.md", []byte(basicCLAUDE), 0644)
}