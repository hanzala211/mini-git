# Mini-Git

A simple Git clone built in Go. We're building this to understand how version control systems work under the hood.

## Installation

Get it installed with:

```bash
go install github.com/hanzalaoc211/mini-git@latest
```

## The `mini-git` Command

Everything starts with the `mini-git` command. It's built using Cobra, so it follows a similar structure to Git with subcommands. Right now we have `init`, `add`, and `commit`.

## What We've Built So Far

### Repository Management

When you run `mini-git init`, it sets up a `.minigit` folder in your project. Inside, we create:

- An `objects` directory where we store all your files (compressed and hashed)
- An `index.json` file that acts as our staging area
- Basic branch references with a `HEAD` file pointing to `master`

### How We Store Files

Files get turned into blob objects using SHA1 hashing, then compressed with zlib before being saved. We store them in a two-level directory structure (`objects/XX/YYYY...`) just like Git does. This means if you add the same file twice, we only store it once - the hash tells us it's already there.

### Staging Files

The `add` command lets you stage files (or entire directories) for commit. It reads the file, creates a blob object, and updates the `index.json` file with the file path and its hash. I skip the `.minigit` folder automatically so you don't accidentally version control your version control files.

### Committing Changes

The `commit` command creates a commit object from the staged files in the index. You must provide a commit message using the `--m` flag. The command builds a tree structure from the staged files, creates a commit object that references the tree, parent commit (if any), commit message, and timestamp. After creating the commit, it updates HEAD to point to the new commit SHA.

Example usage:

```bash
mini-git commit -m "Initial commit"
```

### Working

- **Initialization**: Create a new repository with a `.minigit` directory structure
- **Object Storage**: Files are stored as compressed (zlib) blob objects with SHA1 hashing
- **Index System**: A JSON-based staging area that tracks files and their object hashes
- **Tree Objects**: Directory structures are represented as tree objects that reference blob and other tree objects
- **Commit Objects**: Commits store references to tree objects, parent commits, commit messages, and timestamps
- **Branch References**: Basic branch reference system with HEAD tracking that updates on each commit
