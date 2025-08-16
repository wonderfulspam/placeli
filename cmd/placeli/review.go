package main

import (
	"github.com/spf13/cobra"
	"github.com/user/placeli/internal/tui"
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Review and edit places interactively",
	Long:  "Launch an interactive review interface for detailed place management, editing notes and tags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.RunReview(db)
	},
}

func init() {
	rootCmd.AddCommand(reviewCmd)
}
