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
	rootCmd.AddCommand(EmojiSlideshowCmd)
	rootCmd.AddCommand(FireCmd)
	rootCmd.AddCommand(ClockCmd)
	rootCmd.AddCommand(ShowgifCmd)
	rootCmd.AddCommand(ShowimageCmd)
	rootCmd.AddCommand(TextCmd)
	rootCmd.AddCommand(SnakeCmd)
}
