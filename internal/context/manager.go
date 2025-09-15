package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"baton/internal/llm"
)

// Manager handles context file generation and management for Claude Code
type Manager struct {
	llmClient llm.Client
	workspaceDir string
}

// ProjectContext contains all context information for a project
type ProjectContext struct {
	Name         string
	Vision       string
	Architecture string
	TechStack    []string
	Requirements []string
	Constraints  []string
}

// New creates a new context manager
func New(llmClient llm.Client, workspaceDir string) *Manager {
	return &Manager{
		llmClient:    llmClient,
		workspaceDir: workspaceDir,
	}
}

// GenerateAllContext creates comprehensive context files for Claude Code
func (m *Manager) GenerateAllContext(projectContext *ProjectContext) error {
	// Create context directory structure
	if err := m.createDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Generate core context files
	if err := m.generateCLAUDEFile(projectContext); err != nil {
		return fmt.Errorf("failed to generate CLAUDE.md: %w", err)
	}

	if err := m.generatePRDFiles(projectContext); err != nil {
		return fmt.Errorf("failed to generate PRD files: %w", err)
	}

	if err := m.generateArchitectureDoc(projectContext); err != nil {
		return fmt.Errorf("failed to generate architecture doc: %w", err)
	}

	if err := m.generateStyleGuide(projectContext); err != nil {
		return fmt.Errorf("failed to generate style guide: %w", err)
	}

	if err := m.generateClaudeIgnore(projectContext); err != nil {
		return fmt.Errorf("failed to generate .claudeignore: %w", err)
	}

	if err := m.generateTestingDoc(projectContext); err != nil {
		return fmt.Errorf("failed to generate testing doc: %w", err)
	}

	return nil
}

// createDirectoryStructure sets up the necessary directories
func (m *Manager) createDirectoryStructure() error {
	dirs := []string{
		".claude",
		".claude/subagents",
		"claudedocs",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(m.workspaceDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return err
		}
	}

	return nil
}

// generateCLAUDEFile creates the main CLAUDE.md context file
func (m *Manager) generateCLAUDEFile(projectContext *ProjectContext) error {
	prompt := fmt.Sprintf(`Generate a comprehensive CLAUDE.md file for a project with the following context:

Project: %s
Vision: %s
Architecture: %s
Tech Stack: %s
Requirements: %s
Constraints: %s

Create a CLAUDE.md file that includes:
1. Project overview and purpose
2. Architecture overview with key components
3. Tech stack and framework choices
4. Coding standards and conventions
5. Key directories and file organization
6. Common commands and development workflow
7. Testing approach and requirements
8. Deployment and build process
9. Key patterns to follow and avoid

Format as a complete markdown file that Claude Code can use as comprehensive project context.
Focus on being specific and actionable for an AI assistant working on this codebase.`,
		projectContext.Name,
		projectContext.Vision,
		projectContext.Architecture,
		strings.Join(projectContext.TechStack, ", "),
		strings.Join(projectContext.Requirements, ", "),
		strings.Join(projectContext.Constraints, ", "))

	content, err := m.llmClient.GenerateText(prompt)
	if err != nil {
		return err
	}

	claudePath := filepath.Join(m.workspaceDir, "CLAUDE.md")
	return os.WriteFile(claudePath, []byte(content), 0644)
}

// generatePRDFiles creates detailed product requirements documentation
func (m *Manager) generatePRDFiles(projectContext *ProjectContext) error {
	prompt := fmt.Sprintf(`Generate comprehensive Product Requirements Documentation (PRD) for:

Project: %s
Vision: %s
Requirements: %s

Create a detailed PRD markdown file that includes:
1. Executive Summary
2. Problem Statement
3. User Personas and Use Cases
4. Functional Requirements (detailed with acceptance criteria)
5. Non-Functional Requirements (performance, security, scalability)
6. User Stories and Scenarios
7. Success Metrics and KPIs
8. Technical Constraints
9. Risk Assessment and Mitigations

Format as professional PRD that serves as the single source of truth for product requirements.
Include specific, measurable, and testable requirements that developers can implement against.`,
		projectContext.Name,
		projectContext.Vision,
		strings.Join(projectContext.Requirements, ", "))

	content, err := m.llmClient.GenerateText(prompt)
	if err != nil {
		return err
	}

	prdPath := filepath.Join(m.workspaceDir, "PRD.md")
	return os.WriteFile(prdPath, []byte(content), 0644)
}

// generateArchitectureDoc creates technical architecture documentation
func (m *Manager) generateArchitectureDoc(projectContext *ProjectContext) error {
	prompt := fmt.Sprintf(`Generate comprehensive technical architecture documentation for:

Project: %s
Architecture: %s
Tech Stack: %s

Create detailed ARCHITECTURE.md that includes:
1. System Overview and High-Level Design
2. Component Architecture with relationships
3. Data Architecture and Database Design
4. API Design and Integration Patterns
5. Security Architecture and Considerations
6. Deployment Architecture and Infrastructure
7. Scalability and Performance Considerations
8. Technology Choices and Justifications
9. Architecture Decision Records (ADRs)
10. Development and Deployment Workflows

Format as technical documentation that developers can follow for implementation.
Include diagrams in ASCII art or mermaid format where helpful.`,
		projectContext.Name,
		projectContext.Architecture,
		strings.Join(projectContext.TechStack, ", "))

	content, err := m.llmClient.GenerateText(prompt)
	if err != nil {
		return err
	}

	archPath := filepath.Join(m.workspaceDir, "ARCHITECTURE.md")
	return os.WriteFile(archPath, []byte(content), 0644)
}

// generateStyleGuide creates coding standards and style guide
func (m *Manager) generateStyleGuide(projectContext *ProjectContext) error {
	prompt := fmt.Sprintf(`Generate comprehensive coding style guide for:

Project: %s
Tech Stack: %s

Create detailed STYLE_GUIDE.md that includes:
1. Code Formatting and Conventions
2. Naming Conventions (files, functions, variables, classes)
3. File Organization and Directory Structure
4. Comment and Documentation Standards
5. Error Handling Patterns
6. Testing Patterns and Standards
7. Performance Best Practices
8. Security Best Practices
9. Code Review Guidelines
10. Do's and Don'ts with Examples

Make it specific to the tech stack and include concrete examples.
Focus on patterns that should be consistently followed across the codebase.`,
		projectContext.Name,
		strings.Join(projectContext.TechStack, ", "))

	content, err := m.llmClient.GenerateText(prompt)
	if err != nil {
		return err
	}

	stylePath := filepath.Join(m.workspaceDir, "STYLE_GUIDE.md")
	return os.WriteFile(stylePath, []byte(content), 0644)
}

// generateClaudeIgnore creates .claudeignore file to exclude irrelevant files
func (m *Manager) generateClaudeIgnore(projectContext *ProjectContext) error {
	prompt := fmt.Sprintf(`Generate a .claudeignore file for a project with:

Tech Stack: %s

Create a comprehensive .claudeignore that excludes:
1. Dependencies and package directories
2. Build artifacts and generated files
3. Log files and temporary files
4. IDE and editor files
5. OS-specific files
6. Large data files and media
7. Configuration files with secrets
8. Test coverage reports
9. Documentation build outputs

Format as a gitignore-style file with comments explaining each section.
Be specific to the tech stack being used.`,
		strings.Join(projectContext.TechStack, ", "))

	content, err := m.llmClient.GenerateText(prompt)
	if err != nil {
		return err
	}

	ignorePath := filepath.Join(m.workspaceDir, ".claudeignore")
	return os.WriteFile(ignorePath, []byte(content), 0644)
}

// generateTestingDoc creates testing documentation
func (m *Manager) generateTestingDoc(projectContext *ProjectContext) error {
	prompt := fmt.Sprintf(`Generate comprehensive testing documentation for:

Project: %s
Tech Stack: %s

Create detailed TESTING.md that includes:
1. Testing Strategy and Philosophy
2. Test Types and Coverage Requirements
3. Testing Frameworks and Tools Setup
4. Unit Testing Patterns and Examples
5. Integration Testing Approach
6. End-to-End Testing Strategy
7. Test Data Management
8. Mocking and Stubbing Patterns
9. Performance Testing Guidelines
10. Testing Commands and CI/CD Integration

Make it specific to the tech stack and include executable examples.
Focus on practical guidance developers can immediately use.`,
		projectContext.Name,
		strings.Join(projectContext.TechStack, ", "))

	content, err := m.llmClient.GenerateText(prompt)
	if err != nil {
		return err
	}

	testPath := filepath.Join(m.workspaceDir, "TESTING.md")
	return os.WriteFile(testPath, []byte(content), 0644)
}

// UpdateContext refreshes context files as project evolves
func (m *Manager) UpdateContext(projectContext *ProjectContext) error {
	// Add timestamp to indicate when context was last updated
	projectContext.Name = fmt.Sprintf("%s (updated %s)",
		strings.TrimSuffix(projectContext.Name, fmt.Sprintf(" (updated %s)", time.Now().Format("2006-01-02"))),
		time.Now().Format("2006-01-02"))

	return m.GenerateAllContext(projectContext)
}