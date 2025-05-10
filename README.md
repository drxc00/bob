# 📦 Sweepy

> **Saving space using GO** - A lightweight CLI tool that scans your device for bloat and helps you clean it up.

![Demo Gif](/static/demo.gif)

Sweepy helps developers reclaim valuable disk space by identifying and removing unused `node_modules` directories that accumulate over time from abandoned projects. It's a simple, fast, and efficient tool that can help you save precious space on your device.

![License](https://img.shields.io/github/license/drxc00/sweepy)

## 🚀 Features

- **Fast scanning**: Quickly identifies all `node_modules` directories in your system. Go is just better.
- **Staleness detection**: Analyzes directory staleness such as last modification date.
- **Space visualization**: Shows size statistics to help prioritize cleanup
- **Caching**: Remembers previous scans for improved performance
- **Interactive UI**: Clean TUI interface for easy navigation and cleanup

## 🔧 Installation

```bash
# Clone the repository
git clone https://github.com/drxc00/sweepy.git

# Navigate to the project directory
cd sweepy

# Build the project
go build
```

## 📝 Usage

```bash
sweepy [directory] [flags]

Flags:
  -h, --help                help for scan
  -s, --staleness           The staleness of the scan (default "0")
  -c, --no-cache            Perform a scan without the use of the cache
  -r, --reset-cach          Resets the cache when scanning
  -v, --verbose             Verbose output
```

#### Examples

```bash
# Scan current directory
sweepy

# Scan a specific directory
sweepy "D:\Projects"

# Find node_modules directories not modified in the last 30 days
sweepy "D:\Projects" -s 30

# Perform a fresh scan without using cached results
sweepy "D:\Projects" --no-cache

# Reset the cache and perform a new scan
sweepy "D:\Projects" --reset-cache

# Show detailed progress during scanning
sweepy "D:\Projects" --verbose

```


## 🛠️ Development
This project was created as a learning exercise for Go. Contributions and suggestions for improvements are welcome!

### Roadmap
- Git integration (branches to clean, etc.)
- Comprehensive test suite
- Support for other types of development artifacts (build directories, etc.)
