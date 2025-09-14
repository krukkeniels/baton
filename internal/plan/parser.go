package plan

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"baton/internal/storage"
)

// Parser handles plan file parsing and requirement extraction
type Parser struct{}

// NewParser creates a new plan parser
func NewParser() *Parser {
	return &Parser{}
}

// Plan represents the parsed plan structure
type Plan struct {
	Content      string                `json:"content"`
	Title        string                `json:"title"`
	Sections     map[string]string     `json:"sections"`
	Requirements []*storage.Requirement `json:"requirements"`
}

// Parse parses a markdown plan file
func (p *Parser) Parse(filepath string) (*Plan, []*storage.Requirement, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open plan file: %w", err)
	}
	defer file.Close()

	// Read the entire file
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to read plan file: %w", err)
	}

	content := strings.Join(lines, "\n")

	plan := &Plan{
		Content:  content,
		Sections: make(map[string]string),
	}

	// Extract title (first # heading)
	plan.Title = p.extractTitle(lines)

	// Parse sections
	plan.Sections = p.parseSections(lines)

	// Extract requirements
	requirements, err := p.extractRequirements(lines)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract requirements: %w", err)
	}

	plan.Requirements = requirements

	return plan, requirements, nil
}

// extractTitle extracts the main title from the plan
func (p *Parser) extractTitle(lines []string) string {
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return "Untitled Plan"
}

// parseSections parses markdown sections
func (p *Parser) parseSections(lines []string) map[string]string {
	sections := make(map[string]string)
	currentSection := ""
	currentContent := []string{}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for section headers (# ## ### etc.)
		if strings.HasPrefix(line, "#") {
			// Save previous section
			if currentSection != "" {
				sections[currentSection] = strings.Join(currentContent, "\n")
			}

			// Start new section
			currentSection = strings.TrimSpace(strings.TrimLeft(line, "#"))
			currentContent = []string{}
		} else if currentSection != "" {
			currentContent = append(currentContent, line)
		}
	}

	// Save last section
	if currentSection != "" {
		sections[currentSection] = strings.Join(currentContent, "\n")
	}

	return sections
}

// extractRequirements extracts requirements from the plan content
func (p *Parser) extractRequirements(lines []string) ([]*storage.Requirement, error) {
	var requirements []*storage.Requirement

	// Patterns to match different requirement formats
	patterns := []*regexp.Regexp{
		// FR-P1: Functional requirement pattern
		regexp.MustCompile(`\*\*([A-Z]{2,3}-[A-Z]?\d+)\*\*:\s*(.+)`),
		// FR-1: Alternative pattern
		regexp.MustCompile(`\*\*([A-Z]{2,3}-\d+)\*\*:\s*(.+)`),
		// **FR-P1**: Another pattern
		regexp.MustCompile(`\*\*([A-Z]{2,3}-[A-Z]?\d+)\*\*:\s*(.+)`),
	}

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		for _, pattern := range patterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) >= 3 {
				key := matches[1]
				title := strings.TrimSpace(matches[2])

				// Determine requirement type
				reqType := p.determineRequirementType(key)

				// Extract additional context from following lines
				text := p.extractRequirementText(lines, lineNum)

				requirement := &storage.Requirement{
					ID:    uuid.New().String(),
					Key:   key,
					Title: title,
					Text:  text,
					Type:  reqType,
				}

				requirements = append(requirements, requirement)
				break
			}
		}
	}

	return requirements, nil
}

// determineRequirementType determines the type of requirement from its key
func (p *Parser) determineRequirementType(key string) string {
	switch {
	case strings.HasPrefix(key, "FR"):
		return "functional"
	case strings.HasPrefix(key, "NFR"):
		return "nonfunctional"
	case strings.HasPrefix(key, "CR"), strings.HasPrefix(key, "CON"):
		return "constraint"
	case strings.HasPrefix(key, "RR"), strings.HasPrefix(key, "RISK"):
		return "risk"
	case strings.HasPrefix(key, "AC"):
		return "acceptance"
	default:
		return "functional"
	}
}

// extractRequirementText extracts additional text for a requirement
func (p *Parser) extractRequirementText(lines []string, startLine int) string {
	var textLines []string

	// Look at the current line and a few following lines
	for i := startLine; i < len(lines) && i < startLine+5; i++ {
		line := strings.TrimSpace(lines[i])

		// Stop at the next requirement or section header
		if i > startLine && (strings.Contains(line, "**") && strings.Contains(line, ":") || strings.HasPrefix(line, "#")) {
			break
		}

		if line != "" {
			textLines = append(textLines, line)
		}
	}

	return strings.Join(textLines, " ")
}

// ValidateRequirements validates parsed requirements
func (p *Parser) ValidateRequirements(requirements []*storage.Requirement) []string {
	var issues []string
	seenKeys := make(map[string]bool)

	for _, req := range requirements {
		// Check for duplicate keys
		if seenKeys[req.Key] {
			issues = append(issues, fmt.Sprintf("Duplicate requirement key: %s", req.Key))
		}
		seenKeys[req.Key] = true

		// Check for empty fields
		if req.Title == "" {
			issues = append(issues, fmt.Sprintf("Requirement %s has empty title", req.Key))
		}

		if req.Text == "" {
			issues = append(issues, fmt.Sprintf("Requirement %s has empty text", req.Key))
		}
	}

	return issues
}