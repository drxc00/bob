# bob (WIP)

bob is a lightweight CLI tool that scans your projects for bloat — like unused node_modules — and helps you clean them up with simple, powerful commands.

_I wanted to learn go so I made this. I'm not a go expert, so there's probably a lot of room for improvement._

## Scan

Scan your development environment for clutter like node_modules folders.
Takes a directory as an argument and display stats about the directory given the flags. If no arguments are provided, it defaults to the current directory.

#### Usage

```
bob scan [directory] [flags]

Flags:
  -h, --help                help for scan
  -s, --staleness           The staleness of the scan (default "0")
  -c, --no-cache            Perform a scan without the use of the cache
  -r, --reset-cach          Resets the cache when scanning
  -v, --verbose             Verbose output
```

#### Examples

```
bob scan                                          # Scans the current directory, staleness defaults to 0
bob scan "<directory>"                            # Scans the specified directory, staleness defaults to 0
bob scan "<directory>" -s 1d                      # Scans the specified directory and sets the staleness to 1 day
bob scan "<directory>" --staleness 1d             # Scans the specified directory and sets the staleness to 1 day
bob scan "<directory>" --staleness 1d -no-cache   # Scans the specified directory and sets the staleness to 1 day and disables caching
bob scan "<directory>" --reset-cache              # Resets the cache for the specified directory

```

This retuns a table of the node_modules directories and their size and staleness.

## TODO Features

- Clean command
- Git integration (branches to clean, etc.)
- Flags for clean command
- Flags for scan command
- Tests
- Documentation
