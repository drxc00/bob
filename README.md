# bob (WIP)

A simple and lightweight CLI tool for cleaning up your development environment.

_I wanted to learn go so I made this. I'm not a go expert, so there's probably a lot of room for improvement._

## Scan

Scan your development environment for clutter like node_modules folders, checks for stale git branches, and more in the future.

#### Usage

```
bob scan [directory] [flags]

Flags:
  -h, --help                help for scan
  -s, --staleness           The staleness of the node_modules directory.
  -n, --node                Scan node_modules directories
  -g, --git                 Scan git repositories
```

#### Examples

```
bob scan                                        # Scans the current directory, staleness defaults to 0, node flag defaults to true
bob scan --git                                  # Scans the current directory, staleness defaults to 0, git flag defaults to true
bob scan --node "<directory>"                   # Scans the specified directory, staleness defaults to 0, node flag defaults to true
bob scan --git "<directory>"                    # Scans the specified directory, staleness defaults to 0, git flag defaults to true
bob scan --node "<directory>" -s 1d             # Scans the specified directory and sets the staleness to 1 day, node flag defaults to true
bob scan --git "<directory>" -s 1d              # Scans the specified directory and sets the staleness to 1 day, git flag defaults to true
bob scan --node "<directory>" --staleness 1d    # Scans the specified directory and sets the staleness to 1 day, node flag defaults to true
bob scan --git "<directory>" --staleness 1d     # Scans the specified directory and sets the staleness to 1 day, git flag defaults to true
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
