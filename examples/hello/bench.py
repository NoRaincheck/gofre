#!/usr/bin/env python3
"""
Benchmark script comparing Python vs Go extensions
"""

import time
import sys

def fibonacci_python(n):
    """Pure Python fibonacci implementation."""
    if n <= 1:
        return n
    return fibonacci_python(n-1) + fibonacci_python(n-2)

def sum_slice_python(data):
    """Pure Python sum implementation."""
    return sum(data)

def benchmark_fibonacci(iterations=100):
    """Benchmark fibonacci function."""
    n = 30
    
    print(f"\nBenchmark: fibonacci({n})")
    print("-" * 40)
    
    # Python benchmark
    start = time.perf_counter()
    for _ in range(iterations):
        fibonacci_python(n)
    python_time = (time.perf_counter() - start) / iterations * 1000
    
    # Go benchmark
    try:
        import goforge_example
        start = time.perf_counter()
        for _ in range(iterations):
            goforge_example.Fibonacci(n)
        go_time = (time.perf_counter() - start) / iterations * 1000
        
        speedup = python_time / go_time
        print(f"  Python: {python_time:.3f} ms")
        print(f"  Go:     {go_time:.3f} ms")
        print(f"  Speedup: {speedup:.1f}x")
    except ImportError:
        print("  Go extension not available")
        print(f"  Python: {python_time:.3f} ms")

def benchmark_sum_slice(iterations=100):
    """Benchmark sum_slice function."""
    size = 1_000_000
    data = list(range(size))
    
    print(f"\nBenchmark: sum_slice({size:,})")
    print("-" * 40)
    
    # Python benchmark
    start = time.perf_counter()
    for _ in range(iterations):
        sum_slice_python(data)
    python_time = (time.perf_counter() - start) / iterations * 1000
    
    # Go benchmark
    try:
        import goforge_example
        import array
        arr = array.array('d', data)
        start = time.perf_counter()
        for _ in range(iterations):
            goforge_example.SumSlice(arr)
        go_time = (time.perf_counter() - start) / iterations * 1000
        
        speedup = python_time / go_time
        print(f"  Python: {python_time:.3f} ms")
        print(f"  Go:     {go_time:.3f} ms")
        print(f"  Speedup: {speedup:.1f}x")
    except ImportError:
        print("  Go extension not available")
        print(f"  Python: {python_time:.3f} ms")

def main():
    print("=" * 40)
    print("GoForge Benchmark Suite")
    print("=" * 40)
    
    iterations = 100
    if len(sys.argv) > 1:
        iterations = int(sys.argv[1])
    
    print(f"Iterations: {iterations}")
    
    benchmark_fibonacci(iterations)
    benchmark_sum_slice(iterations)
    
    print("\n" + "=" * 40)
    print("Benchmark complete!")

if __name__ == "__main__":
    main()
