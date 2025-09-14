package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"baton/internal/cycle"
	"baton/internal/llm"
	"baton/internal/storage"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Execute one cycle",
	Long: `Start executes one cycle: select → transition → analyze/execute → handover → completion handshake → audit → stop.

Each cycle advances exactly one task by one valid state transition.`,
	RunE: runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().Duration("timeout", 0, "timeout for cycle execution")
}

func runStart(cmd *cobra.Command, args []string) error {
	// Get timeout from flags
	timeout, _ := cmd.Flags().GetDuration("timeout")
	if timeout == 0 {
		timeout = time.Duration(globalConfig.Development.CycleTimeboxSeconds) * time.Second
	}

	// Create context with timeout
	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	fmt.Printf("⏱ Starting cycle execution (dry-run: %v)\n", globalConfig.Development.DryRunDefault)

	// Initialize database
	store, err := storage.NewStore(globalConfig.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer store.Close()

	// Initialize LLM client
	llmClient, err := createLLMClient()
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Create cycle engine
	engine := cycle.NewCycleEngine(store, globalConfig, llmClient)

	// Execute the cycle
	result, err := engine.ExecuteCycle(ctx, globalConfig.Development.DryRunDefault)
	if err != nil {
		return fmt.Errorf("cycle execution failed: %w", err)
	}

	// Display results
	printCycleResult(result)

	return nil
}

func createLLMClient() (llm.Client, error) {
	// Create client factory
	factory := llm.NewClientFactory()

	// Register Claude client
	claudeClient := llm.NewClaudeClient(&globalConfig.LLM.Claude, globalConfig.MCPPort)
	factory.Register("claude", claudeClient)

	// Get primary client
	client, exists := factory.Get(globalConfig.LLM.Primary)
	if !exists {
		return nil, fmt.Errorf("primary LLM client '%s' not found", globalConfig.LLM.Primary)
	}

	if !client.IsAvailable() {
		return nil, fmt.Errorf("primary LLM client '%s' is not available", globalConfig.LLM.Primary)
	}

	return client, nil
}

func printCycleResult(result *storage.CycleResult) {
	if result.Success {
		fmt.Printf("✅ Cycle completed successfully\n")
	} else {
		fmt.Printf("❌ Cycle failed\n")
	}

	fmt.Printf("Task ID: %s\n", result.TaskID)
	fmt.Printf("State Transition: %s → %s\n", result.PrevState, result.NextState)
	fmt.Printf("Duration: %v\n", result.Duration.Round(time.Millisecond))

	if len(result.ArtifactsCreated) > 0 {
		fmt.Printf("Artifacts Created: %v\n", result.ArtifactsCreated)
	}

	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
	}
}