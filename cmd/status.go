package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"baton/internal/statemachine"
	"baton/internal/storage"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show workspace status",
	Long:  `Status displays summaries by state, recent cycles, blockers, and pending follow-ups.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().Bool("json", false, "output in JSON format")
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Initialize database
	store, err := storage.NewStore(globalConfig.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer store.Close()

	// Get task selector for status information
	selector := statemachine.NewTaskSelector(store, &globalConfig.Selection)
	status, err := selector.GetTaskStatus()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	// Check for JSON output
	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		data, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Human-readable output
	printStatus(status)

	return nil
}

func printStatus(status map[string]interface{}) {
	fmt.Println("📈 Baton Workspace Status")
	fmt.Println("========================")

	// Total tasks
	totalTasks := status["total_tasks"].(int)
	fmt.Printf("Total Tasks: %d\n", totalTasks)

	// Completed tasks
	completedTasks := status["completed_tasks"].(int)
	if totalTasks > 0 {
		completionRate := float64(completedTasks) / float64(totalTasks) * 100
		fmt.Printf("Completion Rate: %.1f%% (%d/%d)\n", completionRate, completedTasks, totalTasks)
	}

	fmt.Println()

	// By state
	fmt.Println("📋 Tasks by State:")
	byState := status["by_state"].(map[string]int)
	for state, count := range byState {
		if count > 0 {
			fmt.Printf("  %s: %d\n", state, count)
		}
	}

	fmt.Println()

	// Ready tasks
	readyTasks := status["ready_tasks"].([]map[string]interface{})
	if len(readyTasks) > 0 {
		fmt.Printf("✅ Ready Tasks (%d):\n", len(readyTasks))
		for i, task := range readyTasks {
			if i >= 5 { // Limit display to first 5
				fmt.Printf("  ... and %d more\n", len(readyTasks)-5)
				break
			}
			fmt.Printf("  %s: %s (Priority: %v)\n", 
				task["id"], 
				task["title"], 
				task["priority"],
			)
		}
	} else {
		fmt.Println("✅ No ready tasks")
	}

	fmt.Println()

	// Blocked tasks
	blockedTasks := status["blocked_tasks"].([]map[string]interface{})
	if len(blockedTasks) > 0 {
		fmt.Printf("⚠️ Blocked Tasks (%d):\n", len(blockedTasks))
		for i, task := range blockedTasks {
			if i >= 5 { // Limit display to first 5
				fmt.Printf("  ... and %d more\n", len(blockedTasks)-5)
				break
			}
			fmt.Printf("  %s: %s\n    Reason: %s\n", 
				task["id"], 
				task["title"], 
				task["reason"],
			)
		}
	} else {
		fmt.Println("⚠️ No blocked tasks")
	}
}