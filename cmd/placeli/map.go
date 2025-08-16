package main

import (
	"github.com/spf13/cobra"
	"github.com/user/placeli/internal/tui"
)

func init() {
	rootCmd.AddCommand(mapCmd)
}

var mapCmd = &cobra.Command{
	Use:   "map",
	Short: "Show places on an interactive ASCII map",
	Long: `Display all places on an interactive ASCII map using terminal graphics.

The map shows place locations using Unicode symbols and allows you to:
- Navigate with arrow keys or vim-style hjkl keys
- Zoom in/out with +/- keys  
- Fit all places in view with 'f'
- Toggle place labels with 't'
- Refresh data with 'r'

Place symbols are chosen based on categories:
🍽 Restaurants    ☕ Cafes        🍺 Bars
🏨 Hotels        ⛽ Gas Stations  🏥 Hospitals  
🏫 Schools       🏦 Banks        🛍 Shopping
🌳 Parks         📍 Default

Examples:
  placeli map                    # Show interactive map
  placeli map | head -20         # Show static map (first 20 lines)
  
Interactive controls:
  ↑↓←→ or hjkl  - Pan the map
  + or =        - Zoom in
  - or _        - Zoom out
  f             - Fit all places
  t             - Toggle labels
  r             - Refresh data
  q             - Quit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.RunMap(db)
	},
}