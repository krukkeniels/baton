package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"baton/internal/config"
	"baton/internal/llm"
	"baton/internal/storage"
	"baton/internal/web"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the web UI server",
	Long: `Start the web UI server for interactive task management.

The web server provides:
- Modern kanban board view of all tasks
- Real-time updates via WebSockets
- Task detail view with complete history
- LLM-powered task creation and updates via prompts
- Responsive design for desktop and mobile

The server will start on the specified port (default: 3001) and serve both
the API endpoints and the static frontend files.`,
	RunE: runWebServer,
}

var (
	webPort     int
	webDevMode  bool
	webStaticDir string
)

func init() {
	rootCmd.AddCommand(webCmd)

	webCmd.Flags().IntVarP(&webPort, "port", "p", 3001, "Port to run the web server on")
	webCmd.Flags().BoolVar(&webDevMode, "dev", false, "Enable development mode with CORS and verbose logging")
	webCmd.Flags().StringVar(&webStaticDir, "static-dir", "./web/dist", "Directory containing static web files")
}

func runWebServer(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize database
	store, err := storage.NewStore(cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}
	defer store.Close()

	// Initialize LLM client
	llmClient, err := llm.NewClient(cfg.LLM)
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Create web server
	webServer := web.NewServer(store, cfg, llmClient)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		log.Printf("Starting web UI server on port %d", webPort)
		if webDevMode {
			log.Println("Development mode enabled - CORS allowed from localhost:3000")
		}
		errChan <- webServer.Start(webPort)
	}()

	// Wait for shutdown signal or error
	select {
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("web server error: %w", err)
		}
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down gracefully...", sig)
		if err := webServer.Stop(); err != nil {
			log.Printf("Error stopping web server: %v", err)
		}
	}

	log.Println("Web server stopped")
	return nil
}