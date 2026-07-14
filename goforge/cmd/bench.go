package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "Run benchmarks comparing Python vs Go extensions",
	Long:  `Run benchmarks to measure the performance improvement of Go extensions`,
	RunE:  runBench,
}

var iterations int

func init() {
	benchCmd.Flags().IntVarP(&iterations, "iterations", "n", 100, "Number of iterations")
	rootCmd.AddCommand(benchCmd)
}

func runBench(cmd *cobra.Command, args []string) error {
	fmt.Println("Running GoForge benchmarks...")
	fmt.Println("================================")
	
	benchmarks := []struct {
		name     string
		pyFunc   string
		goFunc   string
		input    string
	}{
		{"fibonacci(30)", "fibonacci_py(30)", "fibonacci_go(30)", "30"},
		{"fibonacci(40)", "fibonacci_py(40)", "fibonacci_go(40)", "40"},
		{"sum_slice(1M)", "sum_slice_py(data)", "sum_slice_go(data)", "1000000"},
	}
	
	for _, b := range benchmarks {
		fmt.Printf("\nBenchmark: %s\n", b.name)
		fmt.Println("-" + repeat("-", len(b.name)+12))
		
		pyTime := benchmarkPython(b.pyFunc, b.input, iterations)
		goTime := benchmarkGo(b.goFunc, b.input, iterations)
		
		fmt.Printf("  Python:  %v (%d iterations)\n", pyTime, iterations)
		fmt.Printf("  Go:      %v (%d iterations)\n", goTime, iterations)
		
		if goTime > 0 {
			speedup := float64(pyTime) / float64(goTime)
			fmt.Printf("  Speedup: %.1fx\n", speedup)
		}
	}
	
	fmt.Println("\n================================")
	fmt.Println("Benchmark complete!")
	
	return nil
}

func benchmarkPython(funcName, input string, n int) time.Duration {
	pythonCode := fmt.Sprintf(`
import time

def fibonacci_py(n):
    if n <= 1:
        return n
    return fibonacci_py(n-1) + fibonacci_py(n-2)

def sum_slice_py(n):
    return sum(range(n))

start = time.perf_counter()
for _ in range(%d):
    %s
end = time.perf_counter()
print(f"{(end - start) * 1000:.2f}")
`, n, funcName)
	
	cmd := exec.Command("python3", "-c", pythonCode)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  Python error: %s\n", out)
		return 0
	}
	
	var ms float64
	fmt.Sscanf(string(out), "%f", &ms)
	return time.Duration(ms * float64(time.Millisecond))
}

func benchmarkGo(funcName, input string, n int) time.Duration {
	pkgName := getPackageName()
	if pkgName == "" {
		return 0
	}
	
	pythonCode := fmt.Sprintf(`
import time
import %s

start = time.perf_counter()
for _ in range(%d):
    %s
end = time.perf_counter()
print(f"{(end - start) * 1000:.2f}")
`, pkgName, n, funcName)
	
	cmd := exec.Command("python3", "-c", pythonCode)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  Go error: %s\n", out)
		return 0
	}
	
	var ms float64
	fmt.Sscanf(string(out), "%f", &ms)
	return time.Duration(ms * float64(time.Millisecond))
}

func getPackageName() string {
	pyproject := filepath.Join(".", "pyproject.toml")
	if _, err := os.Stat(pyproject); os.IsNotExist(err) {
		fmt.Println("  Warning: pyproject.toml not found")
		return ""
	}
	
	content, err := os.ReadFile(pyproject)
	if err != nil {
		return ""
	}
	
	lines := splitLines(string(content))
	for i, line := range lines {
		if line == "[project]" {
			for j := i + 1; j < len(lines); j++ {
				if len(lines[j]) > 6 && lines[j][:6] == "name =" {
					name := lines[j][6:]
					name = trimSpaces(name)
					name = trimQuotes(name)
					return name
				}
				if len(lines[j]) > 0 && lines[j][0] == '[' {
					break
				}
			}
		}
	}
	
	return ""
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpaces(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func trimQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
