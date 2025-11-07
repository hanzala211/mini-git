package main

import (
	"github.com/hanzalaoc211/mini-git/commands"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "mini-git",
	Short: "A simple git clone",
	Long: "A mini-clone of git",
}

var initCmd = &cobra.Command{
	Use: "init",
	Short: "Initialize a new git repository",
	Long: "Initialize a new git repository in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		commands.InitCommand(cmd, args)
	},
}

var addCmd = &cobra.Command{
	Use: "add",
	Short: "Add file(s) to the staging area",
	Long: "Add one or more files to the staging area in preparation for a commit.",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		commands.AddCommand(cmd, args)
	},
}

func main() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.Execute()
}