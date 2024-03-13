/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"

	"github.com/ry023/reviewhub/reviewhub"
	"github.com/ry023/reviewhub/runners"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "run",
	Short: "Retrieve all source and notify",
	Run: func(cmd *cobra.Command, args []string) {
		cf, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Fatalf("Failed to load flag: %v", err)
		}

		config, err := reviewhub.NewConfig(cf)
		if err != nil {
			log.Fatalf("Failed to parse config file: %v", err)
		}

		r, err := runners.New(config)
		if err != nil {
			log.Fatalf("Failed to create runner: %v", err)
		}

    if err := r.Run(); err != nil {
			log.Fatalf("Failed to run: %v", err)
    }
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("config", "c", ".config.yaml", "config file path")
}
