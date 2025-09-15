package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SubagentType represents different types of specialized agents
type SubagentType string

const (
	ArchitectAgent  SubagentType = "architect"
	DeveloperAgent  SubagentType = "developer"
	ReviewerAgent   SubagentType = "reviewer"
	TesterAgent     SubagentType = "tester"
	DeployerAgent   SubagentType = "deployer"
	DocumenterAgent SubagentType = "documenter"
)

// SubagentSpec defines the configuration for a subagent
type SubagentSpec struct {
	Name        string
	Description string
	Tools       []string
	Prompt      string
}

// GenerateSubagents creates specialized subagent files based on project context
func (m *Manager) GenerateSubagents(projectContext *ProjectContext) error {
	agents := []SubagentType{
		ArchitectAgent,
		DeveloperAgent,
		ReviewerAgent,
		TesterAgent,
		DeployerAgent,
		DocumenterAgent,
	}

	for _, agentType := range agents {
		spec, err := m.generateSubagentSpec(agentType, projectContext)
		if err != nil {
			return fmt.Errorf("failed to generate %s agent: %w", agentType, err)
		}

		if err := m.writeSubagentFile(spec); err != nil {
			return fmt.Errorf("failed to write %s agent file: %w", agentType, err)
		}
	}

	return nil
}

// generateSubagentSpec creates a specialized agent specification
func (m *Manager) generateSubagentSpec(agentType SubagentType, projectContext *ProjectContext) (*SubagentSpec, error) {
	switch agentType {
	case ArchitectAgent:
		return m.generateArchitectAgent(projectContext)
	case DeveloperAgent:
		return m.generateDeveloperAgent(projectContext)
	case ReviewerAgent:
		return m.generateReviewerAgent(projectContext)
	case TesterAgent:
		return m.generateTesterAgent(projectContext)
	case DeployerAgent:
		return m.generateDeployerAgent(projectContext)
	case DocumenterAgent:
		return m.generateDocumenterAgent(projectContext)
	default:
		return nil, fmt.Errorf("unknown agent type: %s", agentType)
	}
}

// generateArchitectAgent creates an architect specialist
func (m *Manager) generateArchitectAgent(projectContext *ProjectContext) (*SubagentSpec, error) {
	prompt := fmt.Sprintf(`Create a specialized architect agent system prompt for:

Project: %s
Architecture: %s
Tech Stack: %s

Generate a detailed system prompt for a Claude Code subagent that specializes in:
- High-level system design and architecture decisions
- Component design and API specification
- Performance and scalability planning
- Technology choice evaluation
- Architecture documentation and ADRs

The agent should be expert in %s and follow established patterns for %s projects.
Include specific guidelines for this project's architecture approach.

Format as a complete system prompt that will be used in a Claude Code subagent.`,
		projectContext.Name,
		projectContext.Architecture,
		strings.Join(projectContext.TechStack, ", "),
		strings.Join(projectContext.TechStack, ", "),
		projectContext.Architecture)

	systemPrompt, err := m.llmClient.GenerateText(prompt)
	if err != nil {
		return nil, err
	}

	return &SubagentSpec{
		Name:        "architect",
		Description: "Specialized in high-level design, architecture decisions, and system planning",
		Tools:       []string{"Read", "Write", "Glob", "Grep", "Bash"},
		Prompt:      systemPrompt,
	}, nil
}

// generateDeveloperAgent creates a developer specialist
func (m *Manager) generateDeveloperAgent(projectContext *ProjectContext) (*SubagentSpec, error) {
	prompt := fmt.Sprintf(`Create a specialized developer agent system prompt for:

Project: %s
Tech Stack: %s
Requirements: %s

Generate a detailed system prompt for a Claude Code subagent that specializes in:
- Feature implementation and coding
- Following established patterns and conventions
- Writing clean, maintainable code
- Implementing business logic and APIs
- Database integration and data handling

The agent should be expert in %s development patterns and follow the project's coding standards.
Focus on practical implementation skills and code quality.

Format as a complete system prompt that will be used in a Claude Code subagent.`,
		projectContext.Name,
		strings.Join(projectContext.TechStack, ", "),
		strings.Join(projectContext.Requirements, ", "),
		strings.Join(projectContext.TechStack, ", "))

	systemPrompt, err := m.llmClient.GenerateText(prompt)
	if err != nil {
		return nil, err
	}

	return &SubagentSpec{
		Name:        "developer",
		Description: "Specialized in feature implementation, coding, and business logic development",
		Tools:       []string{"Read", "Write", "Edit", "MultiEdit", "Glob", "Grep", "Bash"},
		Prompt:      systemPrompt,
	}, nil
}

// generateReviewerAgent creates a code review specialist
func (m *Manager) generateReviewerAgent(projectContext *ProjectContext) (*SubagentSpec, error) {
	prompt := fmt.Sprintf(`Create a specialized code reviewer agent system prompt for:

Project: %s
Tech Stack: %s

Generate a detailed system prompt for a Claude Code subagent that specializes in:
- Code quality assessment and review
- Security vulnerability identification
- Performance optimization recommendations
- Adherence to coding standards and patterns
- Testing coverage and quality evaluation
- Documentation completeness review

The agent should be expert in %s best practices and security considerations.
Focus on constructive feedback and actionable improvement suggestions.

Format as a complete system prompt that will be used in a Claude Code subagent.`,
		projectContext.Name,
		strings.Join(projectContext.TechStack, ", "),
		strings.Join(projectContext.TechStack, ", "))

	systemPrompt, err := m.llmClient.GenerateText(prompt)
	if err != nil {
		return nil, err
	}

	return &SubagentSpec{
		Name:        "reviewer",
		Description: "Specialized in code review, quality assessment, and security analysis",
		Tools:       []string{"Read", "Glob", "Grep", "Bash"},
		Prompt:      systemPrompt,
	}, nil
}

// generateTesterAgent creates a testing specialist
func (m *Manager) generateTesterAgent(projectContext *ProjectContext) (*SubagentSpec, error) {
	prompt := fmt.Sprintf(`Create a specialized testing agent system prompt for:

Project: %s
Tech Stack: %s

Generate a detailed system prompt for a Claude Code subagent that specializes in:
- Test case design and implementation
- Unit, integration, and e2e test creation
- Test data setup and mocking strategies
- Test automation and CI/CD integration
- Performance and load testing
- Bug reproduction and validation

The agent should be expert in %s testing frameworks and patterns.
Focus on comprehensive test coverage and quality assurance.

Format as a complete system prompt that will be used in a Claude Code subagent.`,
		projectContext.Name,
		strings.Join(projectContext.TechStack, ", "),
		strings.Join(projectContext.TechStack, ", "))

	systemPrompt, err := m.llmClient.GenerateText(prompt)
	if err != nil {
		return nil, err
	}

	return &SubagentSpec{
		Name:        "tester",
		Description: "Specialized in test creation, automation, and quality assurance",
		Tools:       []string{"Read", "Write", "Edit", "Bash", "Glob", "Grep"},
		Prompt:      systemPrompt,
	}, nil
}

// generateDeployerAgent creates a deployment specialist
func (m *Manager) generateDeployerAgent(projectContext *ProjectContext) (*SubagentSpec, error) {
	prompt := fmt.Sprintf(`Create a specialized deployment agent system prompt for:

Project: %s
Tech Stack: %s

Generate a detailed system prompt for a Claude Code subagent that specializes in:
- Deployment automation and CI/CD pipelines
- Infrastructure as Code and environment setup
- Container orchestration and cloud deployment
- Monitoring and observability setup
- Production troubleshooting and maintenance
- Security hardening and compliance

The agent should be expert in %s deployment patterns and cloud platforms.
Focus on reliable, secure, and automated deployment processes.

Format as a complete system prompt that will be used in a Claude Code subagent.`,
		projectContext.Name,
		strings.Join(projectContext.TechStack, ", "),
		strings.Join(projectContext.TechStack, ", "))

	systemPrompt, err := m.llmClient.GenerateText(prompt)
	if err != nil {
		return nil, err
	}

	return &SubagentSpec{
		Name:        "deployer",
		Description: "Specialized in deployment automation, infrastructure, and operations",
		Tools:       []string{"Read", "Write", "Bash", "Glob", "Grep"},
		Prompt:      systemPrompt,
	}, nil
}

// generateDocumenterAgent creates a documentation specialist
func (m *Manager) generateDocumenterAgent(projectContext *ProjectContext) (*SubagentSpec, error) {
	prompt := fmt.Sprintf(`Create a specialized documentation agent system prompt for:

Project: %s

Generate a detailed system prompt for a Claude Code subagent that specializes in:
- Technical documentation creation and maintenance
- API documentation and specification
- User guides and tutorials
- Code comments and inline documentation
- Architecture diagrams and visual documentation
- README and setup instructions

The agent should be expert in technical writing and documentation best practices.
Focus on clear, comprehensive, and maintainable documentation.

Format as a complete system prompt that will be used in a Claude Code subagent.`,
		projectContext.Name)

	systemPrompt, err := m.llmClient.GenerateText(prompt)
	if err != nil {
		return nil, err
	}

	return &SubagentSpec{
		Name:        "documenter",
		Description: "Specialized in technical documentation, guides, and knowledge management",
		Tools:       []string{"Read", "Write", "Edit", "Glob", "Grep"},
		Prompt:      systemPrompt,
	}, nil
}

// writeSubagentFile creates the markdown file for a subagent
func (m *Manager) writeSubagentFile(spec *SubagentSpec) error {
	content := fmt.Sprintf(`---
name: %s
description: %s
tools: %s
---

%s
`, spec.Name, spec.Description, strings.Join(spec.Tools, ", "), spec.Prompt)

	filename := fmt.Sprintf("%s.md", spec.Name)
	filePath := filepath.Join(m.workspaceDir, ".claude", "subagents", filename)

	return os.WriteFile(filePath, []byte(content), 0644)
}

// GetSubagentForTask determines which subagent should handle a specific task
func (m *Manager) GetSubagentForTask(taskType, taskDescription string) SubagentType {
	// Simple classification logic - could be enhanced with LLM classification
	taskLower := strings.ToLower(taskType + " " + taskDescription)

	if strings.Contains(taskLower, "plan") || strings.Contains(taskLower, "design") ||
	   strings.Contains(taskLower, "architect") || strings.Contains(taskLower, "api") {
		return ArchitectAgent
	}

	if strings.Contains(taskLower, "review") || strings.Contains(taskLower, "audit") ||
	   strings.Contains(taskLower, "security") || strings.Contains(taskLower, "quality") {
		return ReviewerAgent
	}

	if strings.Contains(taskLower, "test") || strings.Contains(taskLower, "spec") ||
	   strings.Contains(taskLower, "validate") || strings.Contains(taskLower, "coverage") {
		return TesterAgent
	}

	if strings.Contains(taskLower, "deploy") || strings.Contains(taskLower, "build") ||
	   strings.Contains(taskLower, "ci") || strings.Contains(taskLower, "infra") {
		return DeployerAgent
	}

	if strings.Contains(taskLower, "doc") || strings.Contains(taskLower, "readme") ||
	   strings.Contains(taskLower, "guide") || strings.Contains(taskLower, "comment") {
		return DocumenterAgent
	}

	// Default to developer for implementation tasks
	return DeveloperAgent
}