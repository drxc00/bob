# sweepy (WIP)

sweepy is a lightweight dependency free CLI tool that scans your projects for bloat — like unused node_modules — and helps you clean them up with simple, powerful commands.

_I wanted to learn go so I made this. I'm not a go expert, so there's probably a lot of room for improvement._

## Scan

Scan your development environment for clutter like node_modules folders.
Takes a directory as an argument and display stats about the directory given the flags. If no arguments are provided, it defaults to the current directory.

#### Usage

```
sweepy [directory] [flags]

Flags:
  -h, --help                help for scan
  -s, --staleness           The staleness of the scan (default "0")
  -c, --no-cache            Perform a scan without the use of the cache
  -r, --reset-cach          Resets the cache when scanning
  -v, --verbose             Verbose output
```

#### Examples

```
sweepy                                          # Scans the current directory, staleness defaults to 0
sweepy "<directory>"                            # Scans the specified directory, staleness defaults to 0
sweepy "<directory>" -s 1d                      # Scans the specified directory and sets the staleness to 1 day
sweepy "<directory>" --staleness 1d             # Scans the specified directory and sets the staleness to 1 day
sweepy "<directory>" --staleness 1d -no-cache   # Scans the specified directory and sets the staleness to 1 day and disables caching
sweepy "<directory>" --reset-cache              # Resets the cache for the specified directory

```

This retuns a table of the node_modules directories and their size and staleness.

## TODO Features

- Git integration (branches to clean, etc.)
- Tests
- Documentation
