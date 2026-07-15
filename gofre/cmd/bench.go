package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/grackin/gofre/internal/config"
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
	fmt.Println("Running GoFre benchmarks...")
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
		fmt.Println(strings.Repeat("-", len(b.name)+12))

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
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	cfg, err := config.Load(dir)
	if err != nil {
		fmt.Println("  Warning: pyproject.toml not found or invalid")
		return ""
	}
	return cfg.Project.Name
}
