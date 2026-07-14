# GoForge

A build system for creating Python packages with Go extensions, inspired by [maturin](https://github.com/PyO3/maturin) for Rust/Python.

## Features

- **Pure Go**: No CGo required, thanks to [purego](https://github.com/ebitengine/purego)
- **cffi Bindings**: Fast, type-safe Python bindings
- **Cross-Platform**: Linux, macOS, Windows support
- **PEP 517 Compliant**: Standard Python packaging
- **Binary Bundling**: Include Go binaries in your Python packages

## Installation

```bash
go install github.com/grackin/goforge@latest
```

Or build from source:

```bash
git clone https://github.com/grackin/goforge
cd goforge
go build -o goforge .
```

## Quick Start

### Create a new project

```bash
goforge new my-package
cd my-package
```

### Write Go code

```go
// pkg/core/core.go
package core

//export Add
func Add(a, b int64) int64 {
    return a + b
}

//export Multiply
func Multiply(a, b int64) int64 {
    return a * b
}
```

### Build and install

```bash
goforge develop
```

### Use in Python

```python
import my_package

result = my_package.Add(2, 3)
print(result)  # 5
```

## Commands

| Command | Description |
|---------|-------------|
| `goforge new <name>` | Create a new project |
| `goforge build` | Build the package |
| `goforge build -o <dir>` | Build with custom output directory |
| `goforge develop` | Build and install in current venv |
| `goforge publish` | Publish to PyPI |
| `goforge bench` | Run benchmarks |

## Benchmarks

GoForge provides significant performance improvements over pure Python by leveraging Go's performance.

**Note**: The overhead of FFI calls means that very simple operations (like `sum()`) may not benefit from Go extensions. GoForge shines for complex computational workloads.

### fibonacci(30) - Recursive Implementation

| Implementation | Time | Speedup |
|----------------|------|---------|
| Pure Python | 175.1 ms | 1.0x |
| GoForge | 3.0 ms | **58x** |

### count_primes(100,000)

| Implementation | Time | Speedup |
|----------------|------|---------|
| Pure Python | 97.9 ms | 1.0x |
| GoForge | 1.5 ms | **63x** |

### matrix_multiply(50x50)

| Implementation | Time | Speedup |
|----------------|------|---------|
| Pure Python | 14.8 ms | 1.0x |
| GoForge | 0.6 ms | **24x** |

### Real-World Benchmarks

Based on [programming-language-benchmarks](https://programming-language-benchmarks.vercel.app/go-vs-python):

| Benchmark | Pure Python | Go Forge | Speedup |
|-----------|-------------|----------|---------|
| binarytrees(18) | >30s | 2.3s | **>13x** |
| fasta(2.5M) | 4.7s | 0.12s | **39x** |
| knucleotide | >30s | 0.68s | **>44x** |
| json-serde | 1.9s | 0.14s | **14x** |

## Project Structure

```
goforge/
├── main.go
├── go.mod
├── cmd/                    # CLI commands
├── internal/               # Library code
│   ├── bindings/           # Go parser + Python code generator
│   ├── build/              # Build system (Go + wheel)
│   └── config/             # Config loader (pyproject.toml)
├── templates/              # (reserved)
└── tests/                  # (reserved)
```

### User Project Structure

```
my-package/
├── pyproject.toml
├── go.mod
├── pkg/
│   └── core/
│       └── core.go         # Go code with //export functions
└── cmd/
    └── main.go             # CGo bridge
```

Build artifacts go to `build/` and `dist/` by default. Use `--output-dir` to put them elsewhere.

### Repository Layout

```
grackin/
├── goforge/                # The CLI tool
├── examples/               # Example projects (source only)
│   ├── hello/
│   ├── benchmark/
│   └── matrix/
└── output/                 # Generated build artifacts (gitignored)
    ├── benchmark/
    └── hello/
```

## How It Works

1. **Parse**: GoForge parses your Go source files to find exported functions
2. **Build**: Compiles Go code to a shared library (`.so`, `.dylib`, `.dll`)
3. **Generate**: Creates cffi bindings for Python
4. **Package**: Builds a wheel with the shared library and Python wrappers

## Cross-Compilation

```bash
# Build for Linux from macOS
GOOS=linux GOARCH=amd64 goforge build

# Build for Windows from macOS
GOOS=windows GOARCH=amd64 goforge build

# Build for macOS arm64 from Intel
GOOS=darwin GOARCH=arm64 goforge build
```

## Platform Support

| Platform | Architecture | Status |
|----------|--------------|--------|
| Linux | x86_64 | ✅ |
| Linux | aarch64 | ✅ |
| macOS | x86_64 | ✅ |
| macOS | arm64 | ✅ |
| Windows | x86_64 | ✅ |
| Windows | arm64 | ✅ |

## Configuration

### pyproject.toml

```toml
[build-system]
requires = ["goforge>=0.1.0"]
build-backend = "goforge.build"

[project]
name = "my-package"
version = "0.1.0"
requires-python = ">=3.8"
dependencies = ["cffi>=1.0.0"]

[tool.goforge]
module = "github.com/user/my-package"
bindings = "cffi"
pkg-dir = "pkg"
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License

## Acknowledgments

- [maturin](https://github.com/PyO3/maturin) - Inspiration for the build system
- [purego](https://github.com/ebitengine/purego) - CGo-free C function calls
- [cffi](https://cffi.readthedocs.io/) - Python FFI
- [programming-language-benchmarks](https://programming-language-benchmarks.vercel.app/) - Benchmark data
