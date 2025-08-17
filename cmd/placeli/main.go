package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/user/placeli/internal/database"
)

var (
	dbPath string
	db     *database.DB
)

var rootCmd = &cobra.Command{
	Use:   "placeli",
	Short: "Terminal-based Google Maps list manager",
	Long:  "placeli is a terminal-based tool for managing your Google Maps lists with local, offline-first storage.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip database initialization for commands that don't need it
		cmdName := cmd.Name()
		if cmdName == "help" || cmdName == "version" || cmdName == "completion" {
			return
		}

		// Also skip if help flag is set
		if cmd.Flags().Changed("help") {
			return
		}

		if dbPath == "" {
			homeDir, _ := os.UserHomeDir()
			dbPath = filepath.Join(homeDir, ".placeli", "places.db")
		}

		// Create the directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
			os.Exit(1)
		}

		// Initialize database (will create tables if they don't exist)
		var err error
		db, err = database.New(dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			fmt.Fprintf(os.Stderr, "Database will be created at: %s\n", dbPath)
			os.Exit(1)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if db != nil {
			db.Close()
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of placeli",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("placeli v0.1.0 - Terminal-based Google Maps list manager")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "path to SQLite database file")
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
