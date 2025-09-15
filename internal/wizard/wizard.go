package wizard

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"baton/internal/llm"
	"baton/internal/storage"
)

// Wizard handles the interactive project setup process
type Wizard struct {
	llmClient llm.Client
	reader    *bufio.Reader
}

// ProjectInfo contains basic project information
type ProjectInfo struct {
	Name        string
	Vision      string
	Goals       []string
	Constraints []string
	Timeline    string
	TeamSize    string
}

// Requirements contains project requirements
type Requirements struct {
	Functional    []Requirement
	NonFunctional []Requirement
	Constraints   []string
	Risks         []string
}

// Requirement represents a single requirement
type Requirement struct {
	ID          string
	Title       string
	Description string
	Priority    string
	Category    string
}

// Architecture contains technical architecture details
type Architecture struct {
	Overview      string
	TechStack     []string
	Components    []Component
	Integrations  []string
	Deployment    string
	Considerations []string
}

// Component represents a system component
type Component struct {
	Name         string
	Description  string
	Technologies []string
	Dependencies []string
}

// ProjectPlan represents the complete generated plan
type ProjectPlan struct {
	Content   string
	Metadata  map[string]interface{}
	CreatedAt time.Time
}

// Task represents a generated task
type Task struct {
	ID           string
	Title        string
	Description  string
	MVP          string
	State        storage.State
	Priority     int
	Owner        string
	Tags         []string
	Dependencies []string
	Requirements []string
}

// New creates a new wizard instance
func New(llmClient llm.Client, reader *bufio.Reader) *Wizard {
	return &Wizard{
		llmClient: llmClient,
		reader:    reader,
	}
}

// CollectProjectInfo gathers basic project information
func (w *Wizard) CollectProjectInfo() (*ProjectInfo, error) {
	info := &ProjectInfo{}

	// Project name
	fmt.Print("📝 What is your project name? ")
	name, _ := w.reader.ReadString('\n')
	info.Name = strings.TrimSpace(name)

	// Project vision - this is where LLM helps
	fmt.Print("\n🎯 Describe what you want to build (be as detailed as you like):\n> ")
	description, _ := w.reader.ReadString('\n')

	// Get multiple lines if user wants to provide more detail
	fmt.Println("\n(Press Enter twice when done)")
	var fullDescription strings.Builder
	fullDescription.WriteString(description)

	emptyLines := 0
	for {
		line, _ := w.reader.ReadString('\n')
		if strings.TrimSpace(line) == "" {
			emptyLines++
			if emptyLines >= 1 {
				break
			}
		} else {
			emptyLines = 0
			fullDescription.WriteString(line)
		}
	}

	// Use LLM to expand and clarify the vision
	visionPrompt := fmt.Sprintf(`Based on this project description, generate a clear, compelling project vision statement and identify key goals.

Project Name: %s
Description: %s

Please provide a JSON response with:
{
  "vision": "A clear, one-paragraph vision statement",
  "goals": ["3-5 specific, measurable goals"],
  "suggested_timeline": "Realistic timeline estimate",
  "complexity": "low|medium|high",
  "team_size_recommendation": "solo|small|medium|large"
}

Focus on being specific and actionable.`, info.Name, fullDescription.String())

	response, err := w.llmClient.GenerateText(visionPrompt)
	if err != nil {
		// Fallback to user input
		info.Vision = strings.TrimSpace(fullDescription.String())
		info.Goals = []string{"Define core features", "Build MVP", "Deploy to production"}
		info.Timeline = "3-6 months"
		info.TeamSize = "small"
	} else {
		// Parse LLM response
		var visionData struct {
			Vision                   string   `json:"vision"`
			Goals                    []string `json:"goals"`
			SuggestedTimeline        string   `json:"suggested_timeline"`
			Complexity               string   `json:"complexity"`
			TeamSizeRecommendation   string   `json:"team_size_recommendation"`
		}

		if err := json.Unmarshal([]byte(response), &visionData); err == nil {
			info.Vision = visionData.Vision
			info.Goals = visionData.Goals
			info.Timeline = visionData.SuggestedTimeline
			info.TeamSize = visionData.TeamSizeRecommendation
		}
	}

	// Show generated vision for confirmation
	fmt.Println("\n✨ Generated Project Vision:")
	fmt.Println("─────────────────────────────")
	fmt.Printf("%s\n", info.Vision)
	fmt.Println("\n📊 Key Goals:")
	for i, goal := range info.Goals {
		fmt.Printf("   %d. %s\n", i+1, goal)
	}
	fmt.Printf("\n⏱️  Estimated Timeline: %s\n", info.Timeline)
	fmt.Printf("👥 Team Size: %s\n", info.TeamSize)

	fmt.Print("\n✅ Does this look good? (Y/n) ")
	confirm, _ := w.reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(confirm)) == "n" {
		// Allow manual editing
		fmt.Print("\n📝 Please provide your own vision statement:\n> ")
		customVision, _ := w.reader.ReadString('\n')
		info.Vision = strings.TrimSpace(customVision)
	}

	// Collect constraints
	fmt.Print("\n⚠️  Any specific constraints or limitations? (optional, press Enter to skip):\n> ")
	constraints, _ := w.reader.ReadString('\n')
	if strings.TrimSpace(constraints) != "" {
		info.Constraints = strings.Split(constraints, ",")
		for i := range info.Constraints {
			info.Constraints[i] = strings.TrimSpace(info.Constraints[i])
		}
	}

	return info, nil
}

// CollectRequirements gathers detailed requirements
func (w *Wizard) CollectRequirements(projectInfo *ProjectInfo) (*Requirements, error) {
	reqs := &Requirements{}

	fmt.Println("\n📋 Let me generate comprehensive requirements based on your vision...")

	// Use LLM to generate detailed requirements
	reqPrompt := fmt.Sprintf(`Based on this project, generate comprehensive product requirements.

Project: %s
Vision: %s
Goals: %s
Constraints: %s

Generate detailed requirements in JSON format:
{
  "functional": [
    {
      "id": "FR-1",
      "title": "Clear requirement title",
      "description": "Detailed description with acceptance criteria",
      "priority": "high|medium|low",
      "category": "core|feature|enhancement"
    }
  ],
  "non_functional": [
    {
      "id": "NFR-1",
      "title": "Performance requirement",
      "description": "Specific, measurable criteria",
      "priority": "high|medium|low",
      "category": "performance|security|usability|reliability"
    }
  ],
  "constraints": ["Technical or business constraints"],
  "risks": ["Potential risks and mitigation strategies"]
}

Generate 5-10 functional requirements and 3-5 non-functional requirements.
Focus on being specific, testable, and actionable.
Include modern best practices for the type of application.`,
		projectInfo.Name,
		projectInfo.Vision,
		strings.Join(projectInfo.Goals, ", "),
		strings.Join(projectInfo.Constraints, ", "))

	response, err := w.llmClient.GenerateText(reqPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate requirements: %w", err)
	}

	// Parse requirements
	if err := json.Unmarshal([]byte(response), &reqs); err != nil {
		// Try to extract JSON from response
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}") + 1
		if jsonStart >= 0 && jsonEnd > jsonStart {
			jsonStr := response[jsonStart:jsonEnd]
			if err := json.Unmarshal([]byte(jsonStr), &reqs); err != nil {
				return nil, fmt.Errorf("failed to parse requirements: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse requirements: %w", err)
		}
	}

	// Display generated requirements
	fmt.Println("\n📋 Functional Requirements:")
	fmt.Println("──────────────────────────")
	for _, req := range reqs.Functional {
		fmt.Printf("\n%s: %s\n", req.ID, req.Title)
		fmt.Printf("   Priority: %s | Category: %s\n", req.Priority, req.Category)
		fmt.Printf("   %s\n", req.Description)
	}

	fmt.Println("\n🔧 Non-Functional Requirements:")
	fmt.Println("────────────────────────────────")
	for _, req := range reqs.NonFunctional {
		fmt.Printf("\n%s: %s\n", req.ID, req.Title)
		fmt.Printf("   Priority: %s | Category: %s\n", req.Priority, req.Category)
		fmt.Printf("   %s\n", req.Description)
	}

	if len(reqs.Risks) > 0 {
		fmt.Println("\n⚠️  Identified Risks:")
		fmt.Println("─────────────────────")
		for i, risk := range reqs.Risks {
			fmt.Printf("   %d. %s\n", i+1, risk)
		}
	}

	fmt.Print("\n✅ Accept these requirements? (Y/n) ")
	confirm, _ := w.reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(confirm)) == "n" {
		fmt.Println("\n📝 You can edit the generated plan.md file after initialization.")
	}

	return reqs, nil
}

// CollectArchitecture determines technical architecture
func (w *Wizard) CollectArchitecture(projectInfo *ProjectInfo, requirements *Requirements) (*Architecture, error) {
	arch := &Architecture{}

	fmt.Print("\n🛠️  Do you have a preferred tech stack? (optional, press Enter for AI suggestion):\n> ")
	userStack, _ := w.reader.ReadString('\n')
	userStack = strings.TrimSpace(userStack)

	fmt.Println("\n🏗️  Generating optimal architecture...")

	// Generate architecture recommendation
	archPrompt := fmt.Sprintf(`Based on this project, recommend the optimal technical architecture.

Project: %s
Vision: %s
Requirements Summary: %d functional, %d non-functional
User Preference: %s

Consider modern best practices and generate JSON response:
{
  "overview": "High-level architecture description",
  "tech_stack": ["Primary technologies and frameworks"],
  "components": [
    {
      "name": "Component name",
      "description": "What it does",
      "technologies": ["Specific tech for this component"],
      "dependencies": ["Other components it depends on"]
    }
  ],
  "integrations": ["External services/APIs needed"],
  "deployment": "Deployment strategy and platform",
  "considerations": ["Important technical decisions and trade-offs"]
}

Focus on:
- Scalability and maintainability
- Modern development practices
- Security best practices
- Developer experience
- Cost effectiveness`,
		projectInfo.Name,
		projectInfo.Vision,
		len(requirements.Functional),
		len(requirements.NonFunctional),
		userStack)

	response, err := w.llmClient.GenerateText(archPrompt)
	if err != nil {
		// Fallback to basic architecture
		arch.Overview = "Modular architecture with clear separation of concerns"
		arch.TechStack = strings.Split(userStack, ",")
		if len(arch.TechStack) == 0 || arch.TechStack[0] == "" {
			arch.TechStack = []string{"Go", "React", "PostgreSQL"}
		}
		arch.Deployment = "Container-based deployment"
	} else {
		// Parse architecture
		if err := json.Unmarshal([]byte(response), &arch); err != nil {
			// Try to extract JSON
			jsonStart := strings.Index(response, "{")
			jsonEnd := strings.LastIndex(response, "}") + 1
			if jsonStart >= 0 && jsonEnd > jsonStart {
				jsonStr := response[jsonStart:jsonEnd]
				json.Unmarshal([]byte(jsonStr), &arch)
			}
		}
	}

	// Display architecture
	fmt.Println("\n🏗️  Recommended Architecture:")
	fmt.Println("─────────────────────────────")
	fmt.Printf("%s\n", arch.Overview)

	fmt.Println("\n💻 Tech Stack:")
	for _, tech := range arch.TechStack {
		fmt.Printf("   • %s\n", tech)
	}

	if len(arch.Components) > 0 {
		fmt.Println("\n📦 System Components:")
		for _, comp := range arch.Components {
			fmt.Printf("   • %s: %s\n", comp.Name, comp.Description)
		}
	}

	fmt.Printf("\n🚀 Deployment: %s\n", arch.Deployment)

	fmt.Print("\n✅ Use this architecture? (Y/n) ")
	confirm, _ := w.reader.ReadString('\n')

	return arch, nil
}

// GeneratePlan creates the complete plan.md file using pure LLM generation
func (w *Wizard) GeneratePlan(projectInfo *ProjectInfo, requirements *Requirements, architecture *Architecture) (*ProjectPlan, error) {
	// Create comprehensive prompt for LLM to generate complete plan
	prompt := fmt.Sprintf(`Generate a comprehensive project plan document in markdown format for:

PROJECT CONTEXT:
- Name: %s
- Vision: %s
- Goals: %s
- Timeline: %s
- Team Size: %s
- Constraints: %s

REQUIREMENTS:
Functional Requirements:
%s

Non-Functional Requirements:
%s

ARCHITECTURE:
- Overview: %s
- Tech Stack: %s
- Components: %s
- Deployment: %s

Generate a complete plan.md file with these sections:
1. Project title and vision
2. Clear goals and objectives
3. Detailed product requirements (functional and non-functional)
4. Technical architecture and system design
5. MVP-based implementation roadmap (not phases - organize as MVP 1, MVP 2, etc.)
6. Risk assessment and mitigation strategies
7. Success criteria and metrics

For the MVP roadmap:
- Organize tasks into logical MVP deliverables
- Each MVP should be a complete, deployable increment
- Scale the number of MVPs based on project complexity
- Include realistic timelines and dependencies
- Focus on delivering working software incrementally

Format as professional project documentation that serves as the single source of truth.
Include a footer indicating generation by Baton AI Wizard with timestamp.`,
		projectInfo.Name,
		projectInfo.Vision,
		strings.Join(projectInfo.Goals, ", "),
		projectInfo.Timeline,
		projectInfo.TeamSize,
		strings.Join(projectInfo.Constraints, ", "),
		w.formatRequirementsForPrompt(requirements.Functional),
		w.formatRequirementsForPrompt(requirements.NonFunctional),
		architecture.Overview,
		strings.Join(architecture.TechStack, ", "),
		w.formatComponentsForPrompt(architecture.Components),
		architecture.Deployment)

	// Generate complete plan using LLM
	content, err := w.llmClient.GenerateText(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %w", err)
	}

	plan := &ProjectPlan{
		Content:   content,
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"project_name": projectInfo.Name,
			"timeline":     projectInfo.Timeline,
			"team_size":    projectInfo.TeamSize,
		},
	}

	return plan, nil
}

// formatRequirementsForPrompt formats requirements for inclusion in prompts
func (w *Wizard) formatRequirementsForPrompt(requirements []Requirement) string {
	if len(requirements) == 0 {
		return "None specified"
	}

	var formatted strings.Builder
	for _, req := range requirements {
		formatted.WriteString(fmt.Sprintf("- %s (%s): %s [Priority: %s, Category: %s]\n",
			req.ID, req.Title, req.Description, req.Priority, req.Category))
	}
	return formatted.String()
}

// formatComponentsForPrompt formats architecture components for inclusion in prompts
func (w *Wizard) formatComponentsForPrompt(components []Component) string {
	if len(components) == 0 {
		return "Components to be determined"
	}

	var formatted strings.Builder
	for _, comp := range components {
		formatted.WriteString(fmt.Sprintf("- %s: %s (Technologies: %s)\n",
			comp.Name, comp.Description, strings.Join(comp.Technologies, ", ")))
	}
	return formatted.String()
}

// GenerateTasks creates initial tasks from the plan
func (w *Wizard) GenerateTasks(plan *ProjectPlan) ([]Task, error) {
	fmt.Println("\n🤖 Analyzing requirements and creating tasks...")

	taskPrompt := fmt.Sprintf(`Based on this comprehensive project plan, generate a complete waterfall task breakdown.

Full Project Plan:
%s

Generate comprehensive development tasks organized by MVP in JSON format:
{
  "tasks": [
    {
      "title": "Clear, actionable task title",
      "description": "Detailed description of what needs to be done",
      "mvp": "MVP-1",
      "priority": 1-10,
      "tags": ["relevant", "tags"],
      "requirements": ["FR-1", "NFR-2"],
      "estimated_hours": 1-40,
      "dependencies": []
    }
  ]
}

SCALE TASK COUNT BASED ON PROJECT COMPLEXITY:
- Simple projects: 15-30 tasks across 2-3 MVPs
- Medium projects: 30-80 tasks across 3-4 MVPs
- Complex projects: 80-200+ tasks across 4-6 MVPs

ORGANIZE BY MVPs:
- Each MVP should be a complete, deployable increment
- MVP-1: Core foundation and basic functionality
- MVP-2: Primary user value and key features
- MVP-3: Secondary features and enhancements
- MVP-4+: Advanced features, optimization, scaling

TASK REQUIREMENTS:
- Atomic tasks that can be completed in 1-2 cycles
- Clear acceptance criteria in descriptions
- Proper dependency chains between tasks
- Realistic time estimates
- Link to specific requirements where applicable
- Include setup, development, testing, and deployment tasks
- Cover entire software development lifecycle

Create a COMPLETE waterfall breakdown - don't limit task count artificially.`,
		plan.Content[:min(3000, len(plan.Content))])

	response, err := w.llmClient.GenerateText(taskPrompt)
	if err != nil {
		// Generate default tasks
		return w.generateDefaultTasks(), nil
	}

	// Parse tasks
	var taskData struct {
		Tasks []struct {
			Title          string   `json:"title"`
			Description    string   `json:"description"`
			MVP            string   `json:"mvp"`
			Priority       int      `json:"priority"`
			Tags           []string `json:"tags"`
			Requirements   []string `json:"requirements"`
			EstimatedHours int      `json:"estimated_hours"`
			Dependencies   []string `json:"dependencies"`
		} `json:"tasks"`
	}

	if err := json.Unmarshal([]byte(response), &taskData); err != nil {
		// Try to extract JSON
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}") + 1
		if jsonStart >= 0 && jsonEnd > jsonStart {
			jsonStr := response[jsonStart:jsonEnd]
			if err := json.Unmarshal([]byte(jsonStr), &taskData); err != nil {
				return w.generateDefaultTasks(), nil
			}
		} else {
			return w.generateDefaultTasks(), nil
		}
	}

	// Convert to Task objects
	tasks := make([]Task, 0, len(taskData.Tasks))
	for _, td := range taskData.Tasks {
		task := Task{
			ID:           uuid.New().String(),
			Title:        td.Title,
			Description:  td.Description,
			MVP:          td.MVP,
			State:        storage.ReadyForPlan,
			Priority:     td.Priority,
			Owner:        "unassigned",
			Tags:         td.Tags,
			Dependencies: td.Dependencies,
			Requirements: td.Requirements,
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (w *Wizard) generateDefaultTasks() []Task {
	return []Task{
		{
			ID:          uuid.New().String(),
			Title:       "Set up development environment",
			Description: "Configure local development environment with required tools and dependencies",
			State:       storage.ReadyForPlan,
			Priority:    9,
			Tags:        []string{"setup", "infrastructure"},
		},
		{
			ID:          uuid.New().String(),
			Title:       "Initialize project structure",
			Description: "Create initial project structure and configuration files",
			State:       storage.ReadyForPlan,
			Priority:    9,
			Tags:        []string{"setup", "architecture"},
		},
		{
			ID:          uuid.New().String(),
			Title:       "Design database schema",
			Description: "Design and document the database schema based on requirements",
			State:       storage.ReadyForPlan,
			Priority:    8,
			Tags:        []string{"database", "architecture"},
		},
		{
			ID:          uuid.New().String(),
			Title:       "Implement core data models",
			Description: "Create the core data models and domain entities",
			State:       storage.ReadyForPlan,
			Priority:    8,
			Tags:        []string{"backend", "models"},
		},
		{
			ID:          uuid.New().String(),
			Title:       "Set up CI/CD pipeline",
			Description: "Configure continuous integration and deployment pipeline",
			State:       storage.ReadyForPlan,
			Priority:    7,
			Tags:        []string{"devops", "automation"},
		},
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}