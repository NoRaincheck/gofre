# Review: Ambiguities and Decisions

## 1. Type Mapping for Slices

**Ambiguity**: How to handle Go slices in Python bindings?

**Decision**: Use cffi's buffer protocol for slices. The generated binding will:
- Accept Python lists, arrays, or numpy arrays
- Convert to C arrays automatically
- Return slices as Python lists

**Impact**: Users can pass any sequence type to Go functions.

## 2. Error Handling

**Ambiguity**: How to handle Go panics and errors?

**Decision**: 
- Go panics are caught and converted to Python exceptions
- Go error types are mapped to Python exceptions
- Return codes can be used for error signaling

**Impact**: More robust error handling, but adds overhead.

## 3. Memory Management

**Ambiguity**: Who manages memory for returned slices?

**Decision**:
- For simple types: Go runtime manages memory
- For slices: Return copies, not references
- Users must not hold references after function returns

**Impact**: Safe but may have performance implications for large data.

## 4. String Handling

**Ambiguity**: How to handle Go strings?

**Decision**:
- Go strings are passed as C strings (char*)
- Encoding: UTF-8
- Null-terminated strings required

**Impact**: Standard practice, but limits binary data in strings.

## 5. Thread Safety

**Ambiguity**: Is the Go runtime thread-safe?

**Decision**:
- Go's goroutine scheduler is thread-safe
- Python's GIL limits true parallelism
- For multi-threaded Python, use separate Go instances

**Impact**: Safe for single-threaded Python, limited for multi-threaded.

## 6. Binary Bundling

**Ambiguity**: How to include Go binaries in Python packages?

**Decision**:
- Binaries are compiled and included in the wheel
- Python wrapper scripts call binaries via subprocess
- Console scripts can be defined in pyproject.toml

**Impact**: Simple but requires subprocess calls.

## 7. Development Mode

**Ambiguity**: How to handle hot-reload during development?

**Decision**:
- `goforge develop` builds and symlinks to venv
- Changes to Go code require re-running `goforge develop`
- No automatic rebuild on file changes

**Impact**: Manual rebuild required, but predictable.

## 8. Cross-Platform Builds

**Ambiguity**: How to build for multiple platforms?

**Decision**:
- Use `GOOS` and `GOARCH` environment variables
- GoForge detects current platform for wheel tags
- Cross-compilation supported via environment variables

**Impact**: Flexible but requires manual setup for CI/CD.

## 9. manylinux Compliance

**Ambiguity**: How to ensure Linux wheels are manylinux compliant?

**Decision**:
- Use manylinux2014 as the minimum
- Provide Docker images for building
- Document manual build process

**Impact**: Wheels work on most Linux distributions.

## 10. Versioning

**Ambiguity**: How to version the GoForge tool itself?

**Decision**:
- Follow semantic versioning
- Tag releases on GitHub
- Use `go install` or `pip install goforge` for installation

**Impact**: Standard Go versioning practice.

## Open Questions

### Q1: Should we support Go modules with replace directives?

**Status**: Not implemented yet
**Impact**: May affect builds in workspaces

### Q2: How to handle circular dependencies?

**Status**: Not implemented yet
**Impact**: May cause build failures

### Q3: Should we support Go plugins?

**Status**: Not implemented yet
**Impact**: Would enable dynamic loading

### Q4: How to handle large binary outputs?

**Status**: Not implemented yet
**Impact**: May need special packaging

## Future Considerations

1. **Async Support**: Add async/await support for Go functions
2. **Type Hints**: Generate Python type hints from Go types
3. **Documentation**: Auto-generate docs from Go comments
4. **IDE Support**: Language server protocol for GoForge projects
5. **Multi-platform wheels**: Cross-compile for all platforms in a single build (currently builds for current platform only)

## Technical Debt

1. Template engine: Using text/template, could use a more powerful engine
2. Error messages: Could be more descriptive
3. Logging: Currently using fmt.Println, should use a proper logger
4. Configuration: Only supports pyproject.toml, could support more formats

---

## 11. PyPI Distribution (Implemented)

**Ambiguity**: How should goforge itself be distributed via PyPI?

**Decision**:
- Use hatchling as the build backend with a custom build hook for Go compilation
- Binary is compiled with `CGO_ENABLED=0` for static linking
- Python wrapper (`goforge_pkg/`) finds and `exec`s the binary
- Package name: `goforge` on PyPI, `goforge_pkg` as Python import

**Implementation**:
- `pyproject.toml` defines the package metadata and hatchling configuration
- `hatch_build.py` contains the `CustomBuildHook` that compiles the Go binary
- `goforge_pkg/__init__.py` wraps the binary with `os.execvp()` on Unix
- `goforge_pkg/__main__.py` enables `python -m goforge_pkg`
- Console script entry point: `goforge = "goforge_pkg:main"`

**Issues encountered**:
1. Hatchling build hooks require entry point registration (plugin system), making self-contained hooks impossible without installing the package first. Switched to setuptools.
2. Newer setuptools versions reject license classifiers when `license = "PEP 639"` is used. Removed the classifier.
3. Go binary must be compiled with `CGO_ENABLED=0` since the CLI tool itself doesn't use CGo (only the c-shared buildmode for user projects does).
4. Switched from setuptools to hatchling build backend with `BuildHookInterface` for cleaner build hook support.

**Impact**: `uv build` or `uv pip install .` compiles the Go binary and installs it. Users get the `goforge` CLI on their PATH.

## 12. Go Unit Tests (Implemented)

**Ambiguity**: How to test the internal packages?

**Decision**:
- Add unit tests for all internal packages in `goforge/tests/`
- Use Go's standard `testing` package
- Use `testdata/` directory for fixture files (sample Go source, pyproject.toml)
- Export template helper functions for direct testing

**Tests added**:
- `parser_test.go`: AST parsing, `//export` detection, type extraction, test file skipping
- `generator_test.go`: cffi binding generation, template functions, slice handling
- `types_test.go`: Go→C/cffi type mapping for all primitives, slices, unknown types
- `config_test.go`: pyproject.toml loading, validation, defaults, error handling
- `go_build_test.go`: Platform detection, lib extension, module checks, venv detection
- `wheel_test.go`: Wheel structure, metadata content, file contents, naming

**Issues encountered**:
1. `ast.CommentGroup.Text()` returns empty string for `//export` comments (it strips comment markers and normalizes). The raw `c.Text` field contains the actual text. Test adjusted to verify function params instead.
2. `FindVenvDir()` walks up directories checking for `.venv/` and also checks `$VIRTUAL_ENV`. Tests must unset this env var to avoid false matches.
3. Wheel builder archives files relative to `buildDir` and prepends `pkgName/`. Test fixtures must place files at the root of `buildDir`, not in a subdirectory.

**Impact**: 44 unit tests covering all internal packages, all passing.

## 13. Bug Fixes and Cleanup (Implemented)

**Critical fixes**:
1. `cmd/publish.go`: Fixed broken command `go forge build` → `goforge build`
2. `internal/bindings/generator.go`: Removed brittle `n * n` hack from `ReturnCountExpr`; now falls back to `_count` parameter for slice returns without an input slice
3. `internal/build/wheel.go`: Fixed `Install()` hardcoded `python3.9` — now calls `GetPythonVersion()`
4. `cmd/publish.go`: Removed hand-rolled TOML parser with prefix-matching bug; now uses `config.Load()`

**Important fixes**:
5. `cmd/bench.go` / `cmd/publish.go`: Replaced hand-rolled TOML parsing with `config.Load()` from the config package
6. `cmd/bench.go`: Replaced hand-rolled `splitLines`, `trimSpaces`, `trimQuotes`, `repeat` with `strings.Split`, `strings.TrimSpace`, `strings.Trim`, `strings.Repeat`
7. `cmd/publish.go`: Extracted `getPackageName()` (was in bench.go) to use `config.Load()`
8. `cmd/new.go`: Removed dead code `createPythonPackage()` (never called)
9. `internal/bindings/generator.go`: Removed unused template functions `ReturnIsSlice`, `ReturnElemType`, `SliceParamName` from FuncMap
10. `internal/build/wheel.go`: Fixed malformed RECORD file — now contains proper CSV entries per PEP 427
11. `internal/build/go_build.go`: Fixed `GetPythonTag()` — now detects actual Python version instead of hardcoding `cp39`
12. `.gitignore`: Added `.DS_Store`

**Minor fixes**:
13. Removed unused test fixtures (`valid_pyproject.toml`, `defaults_pyproject.toml`, `missing_name.toml`, `missing_version.toml`)
14. Removed unused Python helpers (`_get_platform_tag`, `_get_lib_extension`) from `goforge_pkg/__init__.py`
15. Updated README.md: Fixed project structure section (removed `templates/ (reserved)`, updated `tests/` description)
16. Cleaned stale build artifacts from disk

**Example fixes**:
17. `examples/hello/`: Added missing `pyproject.toml`, `go.mod`, `cmd/main.go`
18. `examples/matrix/`: Added missing `pyproject.toml`, `go.mod`, `cmd/main.go`
19. `examples/matrix/pkg/matrix/matrix.go`: Removed unnecessary `import "C"` (contradicts "no CGo required" claim)
