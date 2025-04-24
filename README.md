# bob

A lightweight dependency-free CLI tool that helps you keep your development environment clean and clutter-free.

## commands

### scan

```
bob scan                  # Default: scan node_modules
bob scan --all            # Scan all supported types
bob scan --git            # Scan for git clutter (e.g. stale branches, large .git folders)
bob scan --cache          # Scan for known cache folders (.cache, dist, build)
bob scan --node           # Explicitly scan only node_modules (same as default)
bob scan --path ./mydir   # Restrict scanning to a directory
```

### clean

```
bob clean --git           # Clean untracked git files or stale branches (optional, safe mode)
bob clean --node          # Clean node_modules (default)
bob clean --cache
```

### staleness

```
bob scan --node --staleness 30
bob scan --git --staleness 180
bob scan --cache --staleness 90

bob clean --node --staleness 30
bob clean --git --staleness 180
bob clean --cache --staleness 90
```
