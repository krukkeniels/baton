package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"baton/internal/config"
	"baton/pkg/version"
)

var (
	cfgFile    string
	workspace  string
	dryRun     bool
	verbose    bool
	globalConfig *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "baton",
	Short: "CLI Orchestrator for LLM-Driven Task Execution",
	Long: `Baton is a CLI orchestrator that advances work one task state at a time.

Execution is organized into cycles. Each cycle advances exactly one task by one 
valid state transition, with context cleared between cycles and formal handover 
artifacts to bridge cycles.`,
	Version: version.Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.baton/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&workspace, "workspace", "./", "workspace directory")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would be done without making changes")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to viper
	viper.BindPFlag("workspace", rootCmd.PersistentFlags().Lookup("workspace"))
	viper.BindPFlag("dry_run", rootCmd.PersistentFlags().Lookup("dry-run"))
}

// initConfig reads in config file and ENV variables.
func initConfig() {
	var err error
	globalConfig, err = config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Override with command line flags
	if workspace != "" {
		globalConfig.Workspace = workspace
	}

	if dryRun {
		globalConfig.Development.DryRunDefault = true
	}
}