# bob (WIP)

Your personal Bloat Observer and Buster — effortlessly scan, spot, and smash stale files, folders, and Git clutter across your projects.

bob is a lightweight CLI tool that scans your projects for bloat — like unused node_modules, stale Git branches, and others — and helps you clean them up with simple, powerful commands. Built for developers who want faster, tidier, and more maintainable projects without the hassle.

_I wanted to learn go so I made this. I'm not a go expert, so there's probably a lot of room for improvement._

## Scan

Scan your development environment for clutter like node_modules folders and more (in the future).
Takes a directory as an argument and display stats about the directory given the flags. If no arguments are provided, it defaults to the current directory.

#### Usage

```
bob scan [directory] [flags]

Flags:
  -h, --help                help for scan
  -s, --staleness           The staleness of the scan (default "0")
  -n, --node                Scan node_modules directories
  -c, --no-cache            Disable caching
```

#### Examples

```
bob scan --node                                         # Scans the current directory, staleness defaults to 0, git flag defaults to true
bob scan --node "<directory>"                           # Scans the specified directory, staleness defaults to 0, node flag defaults to true
bob scan --node "<directory>" -s 1d                     # Scans the specified directory and sets the staleness to 1 day, node flag defaults to true
bob scan --node "<directory>" --staleness 1d            # Scans the specified directory and sets the staleness to 1 day, node flag defaults to true
bob scan --node "<directory>" --staleness 1d -no-cache  # Scans the specified directory and sets the staleness to 1 day, node flag defaults to true, and disables caching

```

This retuns a table of the node_modules directories and their size and staleness.

## Clean

TODO

## TODO Features

- Clean command
- Git integration (branches to clean, etc.)
- Flags for clean command
- Flags for scan command
- Tests
- Documentation
