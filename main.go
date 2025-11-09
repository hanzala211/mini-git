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

var commitCmd = &cobra.Command{
	Use: "commit",
	Short: "Commit the staged changes",
	Long: "Commit the staged changes to the repository",
	Run: func(cmd *cobra.Command, args []string) {
		commands.CommitCommand(cmd, args)
	},
}

var branchCmd = &cobra.Command{
	Use: "branch",
	Short: "List and Create New Branches",
	Long: "List and Create New Branches",
	Run: func(cmd *cobra.Command, args []string) {
		commands.BranchCommand(cmd, args)
	},
}

var checkoutCmd = &cobra.Command{
	Use: "checkout",
	Short: "Switch to a different branch",
	Long: "Switch to a different branch and update the working directory",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		commands.CheckoutCommand(cmd, args)
	},
}

var mergeCmd = &cobra.Command{
	Use: "merge",
	Short: "Merge a branch into the current branch",
	Long: "Merge another branch into the current branch",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		commands.MergeCommand(cmd, args)
	},
}

func main() {

	commitCmd.Flags().String("m", "", "message for the commit (required)")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(branchCmd)
	rootCmd.AddCommand(checkoutCmd)
	rootCmd.AddCommand(mergeCmd)
	rootCmd.Execute()
}