package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/placeli/internal/web"
)

var (
	webPort   int
	webAPIKey string
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start web interface for browsing places",
	Long: `Start a web server that provides an interactive map interface for browsing your saved places.

The web interface includes:
- Interactive map showing all your places
- Search and filter functionality
- Detailed place information
- Mobile-friendly responsive design

You can optionally provide a Google Maps API key for enhanced map features.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if webAPIKey == "" {
			webAPIKey = os.Getenv("GOOGLE_MAPS_API_KEY")
		}

		server, err := web.NewServer(db, webPort, webAPIKey)
		if err != nil {
			return fmt.Errorf("failed to create server: %w", err)
		}

		return server.Start()
	},
}

func init() {
	webCmd.Flags().IntVarP(&webPort, "port", "p", 8080, "port to run web server on")
	webCmd.Flags().StringVar(&webAPIKey, "api-key", "", "Google Maps API key (optional, uses env GOOGLE_MAPS_API_KEY if not set)")

	rootCmd.AddCommand(webCmd)
}
