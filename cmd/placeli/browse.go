package main

import (
	"github.com/spf13/cobra"
	"github.com/user/placeli/internal/tui"
)

var browseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Browse places with interactive TUI",
	Long:  "Launch an interactive terminal UI for browsing and managing your saved places.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.RunBrowse(db)
	},
}

func init() {
	rootCmd.AddCommand(browseCmd)
}
