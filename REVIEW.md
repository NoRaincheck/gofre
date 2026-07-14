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
- Use `go install` for installation

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
5. **Testing**: Integration with pytest and go test

## Technical Debt

1. Template engine: Using text/template, could use a more powerful engine
2. Error messages: Could be more descriptive
3. Logging: Currently using fmt.Println, should use a proper logger
4. Configuration: Only supports pyproject.toml, could support more formats
5. Testing: No unit tests yet
