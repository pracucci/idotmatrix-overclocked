package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "idm-cli",
	Short: "A simple CLI application to interact with iDot displays",
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(DiscoverCmd)
	rootCmd.AddCommand(EmojiCmd)
	rootCmd.AddCommand(DemoCmd)
	rootCmd.AddCommand(FireCmd)
	rootCmd.AddCommand(ClockCmd)
	rootCmd.AddCommand(GrotCmd)
	rootCmd.AddCommand(OffCmd)
	rootCmd.AddCommand(OnCmd)
	rootCmd.AddCommand(ServerCmd)
	rootCmd.AddCommand(ShowgifCmd)
	rootCmd.AddCommand(ShowimageCmd)
	rootCmd.AddCommand(TextCmd)
	rootCmd.AddCommand(SnakeCmd)
}
