package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"baton/internal/plan"
	"baton/internal/storage"
)

// ingestCmd represents the ingest command
var ingestCmd = &cobra.Command{
	Use:   "ingest [plan-file]",
	Short: "Parse plan file and ingest requirements",
	Long: `Ingest parses a markdown plan file and extracts requirements into the database.

This command will:
1. Parse the plan file for requirements (FR-*, NFR-*, etc.)
2. Create or update requirements in the database
3. Report any parsing errors or validation issues

The command is idempotent - running it multiple times will update existing requirements.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runIngest,
}

func init() {
	rootCmd.AddCommand(ingestCmd)
}

func runIngest(cmd *cobra.Command, args []string) error {
	// Determine plan file path
	planFile := globalConfig.PlanFile
	if len(args) > 0 {
		planFile = args[0]
	}

	fmt.Printf("📄 Ingesting plan file: %s\n", planFile)

	// Initialize database
	store, err := storage.NewStore(globalConfig.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer store.Close()

	// Parse plan file
	parser := plan.NewParser()
	parsedPlan, requirements, err := parser.Parse(planFile)
	if err != nil {
		return fmt.Errorf("failed to parse plan file: %w", err)
	}

	fmt.Printf("Plan Title: %s\n", parsedPlan.Title)
	fmt.Printf("Found %d requirements\n", len(requirements))

	// Validate requirements
	issues := parser.ValidateRequirements(requirements)
	if len(issues) > 0 {
		fmt.Println("\n⚠️ Validation Issues:")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
		fmt.Println()
	}

	// Ingest requirements
	var created, updated int
	for _, req := range requirements {
		// Check if requirement already exists
		existing, err := store.GetRequirement(req.Key)
		if err != nil {
			// Doesn't exist, create new
			if err := store.CreateRequirement(req); err != nil {
				fmt.Printf("❌ Failed to create requirement %s: %v\n", req.Key, err)
				continue
			}
			created++
			fmt.Printf("✅ Created: %s\n", req.Key)
		} else {
			// Exists, update if different
			if existing.Title != req.Title || existing.Text != req.Text || existing.Type != req.Type {
				// Update existing requirement
				existing.Title = req.Title
				existing.Text = req.Text
				existing.Type = req.Type

				if err := store.UpdateRequirement(existing); err != nil {
					fmt.Printf("❌ Failed to update requirement %s: %v\n", req.Key, err)
					continue
				}
				updated++
				fmt.Printf("🔄 Updated: %s\n", req.Key)
			} else {
				fmt.Printf("✔️ No changes: %s\n", req.Key)
			}
		}
	}

	fmt.Printf("\n📈 Ingestion Summary:\n")
	fmt.Printf("  Created: %d requirements\n", created)
	fmt.Printf("  Updated: %d requirements\n", updated)
	fmt.Printf("  Total: %d requirements\n", len(requirements))

	if len(issues) == 0 {
		fmt.Println("✅ Plan ingestion completed successfully!")
	} else {
		fmt.Printf("⚠️ Plan ingestion completed with %d validation issues\n", len(issues))
	}

	return nil
}