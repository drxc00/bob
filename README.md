# bob

A simple and lightweight CLI tool for cleaning up your development environment.

## Scan

Scan your development environment for clutter like node_modules folders.
Returns a list of node_modules directories and their size and staleness (how long since the last time the directory was modified).
This information is useful for identifying projects that haven't been updated in a while. We can then remove the node_modules folder to free up space.
We can always add the node_modules folder back if we need it.

#### Usage

```
bob scan [directory] [flags]

Flags:
  -h, --help                help for scan
  -s, --staleness string    The staleness of the node_modules directory (days, hrs, mins, secs)
  -n, --node                Scan node_modules directories
  -g, --git                 Scan git repositories
```

#### Examples

```
bob scan                                    # Scans the current directory (staleness defaults to 0)
bob scan "<directory>"                      # Scans the specified directory (staleness defaults to 0)
bob scan "<directory>" -s 1d                # Scans the specified directory and sets the staleness to 1 day
bob scan "<directory>" --staleness 1d       # Scans the specified directory and sets the staleness to 1 day
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
