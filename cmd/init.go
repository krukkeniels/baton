package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"baton/internal/config"
	"baton/internal/storage"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Baton workspace",
	Long: `Initialize creates a new Baton workspace with default configuration,
database, and sample plan file.

This command is idempotent and safe to run multiple times.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	workspaceDir := globalConfig.Workspace

	fmt.Printf("Initializing Baton workspace in: %s\n", workspaceDir)

	// Create workspace directory
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return fmt.Errorf("failed to create workspace directory: %w", err)
	}

	// Create default config if it doesn't exist
	configPath := filepath.Join(workspaceDir, "baton.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("Creating default configuration: %s\n", configPath)
		if err := config.CreateDefaultConfig(configPath); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
	} else {
		fmt.Printf("Configuration already exists: %s\n", configPath)
	}

	// Initialize database
	dbPath := filepath.Join(workspaceDir, "baton.db")
	fmt.Printf("Initializing database: %s\n", dbPath)
	
	store, err := storage.NewStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer store.Close()

	// Create sample plan file if it doesn't exist
	planPath := filepath.Join(workspaceDir, "plan.md")
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		fmt.Printf("Creating sample plan file: %s\n", planPath)
		if err := createSamplePlan(planPath); err != nil {
			return fmt.Errorf("failed to create sample plan: %w", err)
		}
	} else {
		fmt.Printf("Plan file already exists: %s\n", planPath)
	}

	fmt.Println("\n✅ Baton workspace initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Printf("1. Edit your plan file: %s\n", planPath)
	fmt.Printf("2. Ingest the plan: baton ingest %s\n", planPath)
	fmt.Println("3. Start your first cycle: baton start")

	return nil
}

func createSamplePlan(path string) error {
	samplePlan := `# Sample Project Plan

## Vision
This is a sample project to demonstrate Baton's capabilities.

## Scope
Implement a simple task management system.

## Functional Requirements

**FR-P1**: The system shall allow users to create new tasks.

**FR-P2**: The system shall allow users to view existing tasks.

**FR-P3**: The system shall allow users to mark tasks as completed.

**FR-P4**: The system shall persist tasks in a database.

## Non-Functional Requirements

**NFR-1**: The system shall respond to user actions within 200ms.

**NFR-2**: The system shall be available 99.9% of the time.

## Constraints

**CR-1**: The system must use SQLite for data persistence.

**CR-2**: The system must have a REST API.

## Roadmap

### Phase 1: Core Functionality
- Implement basic task CRUD operations
- Set up database schema
- Create REST API endpoints

### Phase 2: User Interface
- Create web-based UI
- Implement task filtering
- Add user authentication

### Phase 3: Advanced Features
- Task categories and tags
- Due dates and reminders
- Task sharing and collaboration
`

	return os.WriteFile(path, []byte(samplePlan), 0644)
}