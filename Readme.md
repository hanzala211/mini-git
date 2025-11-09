# Mini-Git

A simple Git clone built in Go. I'm building this to understand how version control systems work under the hood.

## Installation

Get it installed with:

```bash
go install github.com/hanzalaoc211/mini-git@latest
```

## The `mini-git` Command

Everything starts with the `mini-git` command. It's built using Cobra, so it follows a similar structure to Git with subcommands. Right now I have `init`, `add`, `commit`, `branch`, `checkout`, and `merge`.

## What I've Built So Far

### Repository Management

When you run `mini-git init`, it sets up a `.minigit` folder in your project. Inside, I create:

- An `objects` directory where I store all your files (compressed and hashed)
- An `index.json` file that acts as my staging area
- Basic branch references with a `HEAD` file pointing to `master`

### How I Store Files

Files get turned into blob objects using SHA1 hashing, then compressed with zlib before being saved. I store them in a two-level directory structure (`objects/XX/YYYY...`) just like Git does. This means if you add the same file twice, I only store it once - the hash tells me it's already there.

### Staging Files

The `add` command lets you stage files (or entire directories) for commit. It reads the file, creates a blob object, and updates the `index.json` file with the file path and its hash. I skip the `.minigit` folder automatically so you don't accidentally version control your version control files.

### Committing Changes

The `commit` command creates a commit object from the staged files in the index. You must provide a commit message using the `--m` flag. The command builds a tree structure from the staged files, creates a commit object that references the tree, parent commit (if any), commit message, and timestamp. After creating the commit, it updates HEAD to point to the new commit SHA.

Example usage:

```bash
mini-git commit --m "Initial commit"
```

### Branching

The `branch` command lets you create and list branches. When you run `mini-git branch` without arguments, it lists all available branches with an asterisk marking the current one. When you provide a branch name, it creates a new branch pointing to the current commit and automatically switches to it.

Example usage:

```bash
# List all branches
mini-git branch

# Create and switch to a new branch
mini-git branch feature-branch
```

### Checkout (This is a Must!)

The `checkout` command is essential for switching between branches. When you switch branches, it compares the tree objects of the current branch and the target branch, then updates your working directory accordingly. It handles:

- **Modified files**: Files that changed between branches get updated
- **New files**: Files that exist in the target branch but not in the current one get created
- **Deleted files**: Files that exist in the current branch but not in the target one get removed
- **Unchanged files**: Files with the same SHA are left untouched (smart, right?)

The command also updates the HEAD reference to point to the new branch. If you try to checkout the branch you're already on, it just tells you that you're already there.

Example usage:

```bash
mini-git checkout feature-branch
```

### Merge

The `merge` command combines changes from one branch into the current branch. Currently, it supports **fast-forward merges** only. The implementation (see `commands/merge.go`) works as follows:

When you merge a branch, it first checks if the current branch is an ancestor of the branch being merged. If it is, a fast-forward merge is performed:

1. **Ancestor Check**: Uses the `isAncestor` function to traverse the commit history and determine if the current branch's commit is an ancestor of the branch being merged
2. **Branch Update**: Updates the current branch reference to point to the merged branch's commit
3. **Working Directory Update**: Uses `diffAndApply` to compare tree objects and update files in your working directory:
   - Files that changed get updated
   - New files from the merged branch get created
   - Files removed in the merged branch get deleted
   - Unchanged files remain untouched
4. **Index Update**: Rebuilds the index from the merged branch's tree using `buildIndexFromTree` to reflect the new state

If you try to merge the branch you're already on, it will inform you that you're already on that branch. If the merge requires a 3-way merge (when branches have diverged), it will report that 3-way merges are not yet implemented.

Example usage:

```bash
# Merge feature-branch into the current branch
mini-git merge feature-branch
```

### What's Working

- **Initialization**: Create a new repository with a `.minigit` directory structure
- **Object Storage**: Files are stored as compressed (zlib) blob objects with SHA1 hashing
- **Index System**: A JSON-based staging area that tracks files and their object hashes
- **Tree Objects**: Directory structures are represented as tree objects that reference blob and other tree objects
- **Commit Objects**: Commits store references to tree objects, parent commits, commit messages, and timestamps
- **Branch References**: Branch reference system with HEAD tracking that updates on each commit
- **Branch Management**: Create and list branches, with automatic switching on creation
- **Checkout**: Switch between branches with intelligent working directory updates
- **Merge**: Fast-forward merge support that combines branches when the current branch is an ancestor of the merged branch

## What's Next

I'm planning to add `status` and `log` commands next.
