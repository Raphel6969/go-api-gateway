package main

import (
	"fmt"
	"os"

	"github.com/Raphel6969/api-gateway/internal/config"
	"github.com/spf13/cobra"
)

var valCfgFile string

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validates the gateway configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := config.LoadConfig(valCfgFile)
		if err != nil {
			fmt.Printf("❌ Configuration invalid: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Configuration is valid!")
	},
}

func init() {
	validateCmd.Flags().StringVarP(&valCfgFile, "config", "c", "gateway.yaml", "config file to validate")
	rootCmd.AddCommand(validateCmd)
}
