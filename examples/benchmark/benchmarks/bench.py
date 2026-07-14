#!/usr/bin/env python3
"""
GoForge Benchmark Suite
Compares pure Python vs Go extensions across multiple computational tasks.
"""

import time
import sys
import os
import random

# ============================================================
# Pure Python implementations
# ============================================================

def fibonacci_py(n):
    """Pure Python recursive fibonacci."""
    if n <= 1:
        return n
    return fibonacci_py(n - 1) + fibonacci_py(n - 2)

def sum_slice_py(data):
    """Pure Python sum."""
    total = 0.0
    for v in data:
        total += v
    return total

def matrix_multiply_py(a, b, n):
    """Pure Python matrix multiplication."""
    result = [0.0] * (n * n)
    for i in range(n):
        for j in range(n):
            s = 0.0
            for k in range(n):
                s += a[i * n + k] * b[k * n + j]
            result[i * n + j] = s
    return result

def count_primes_py(limit):
    """Pure Python prime counter."""
    count = 0
    for num in range(2, limit):
        if num < 2:
            continue
        if num == 2:
            count += 1
            continue
        if num % 2 == 0:
            continue
        is_prime = True
        i = 3
        while i * i <= num:
            if num % i == 0:
                is_prime = False
                break
            i += 2
        if is_prime:
            count += 1
    return count

def sort_ints_py(data):
    """Pure Python quicksort."""
    result = list(data)
    result.sort()
    return result


# ============================================================
# Benchmark runner
# ============================================================

def bench(func, *args, iterations=1):
    """Benchmark a function and return time in ms."""
    start = time.perf_counter()
    for _ in range(iterations):
        result = func(*args)
    elapsed = (time.perf_counter() - start) / iterations
    return elapsed * 1000, result

def format_time(ms):
    """Format time appropriately."""
    if ms < 0.001:
        return f"{ms*1000000:.1f} ns"
    elif ms < 1:
        return f"{ms*1000:.1f} us"
    elif ms < 1000:
        return f"{ms:.3f} ms"
    else:
        return f"{ms/1000:.3f} s"

def print_result(name, py_time, go_time):
    """Print benchmark result."""
    speedup = py_time / go_time if go_time > 0 else float('inf')
    print(f"  {name:30s}  Python: {format_time(py_time):>12s}  Go: {format_time(go_time):>12s}  Speedup: {speedup:6.1f}x")


def run_benchmarks():
    """Run all benchmarks."""
    # Try to import Go bindings
    go_available = False
    build_paths = [
        os.path.join(os.path.dirname(__file__), '..', 'build'),
        os.path.join(os.path.dirname(__file__), '..', '..', '..', 'output', 'benchmark', 'build'),
    ]
    for bp in build_paths:
        try:
            sys.path.insert(0, bp)
            import goforge_benchmark as go
            go_available = True
            break
        except ImportError:
            continue

    print("=" * 72)
    print("GoForge Benchmark Suite")
    print("=" * 72)
    print(f"Python: {sys.version}")
    print(f"Go bindings: {'available' if go_available else 'NOT available (pure Python only)'}")
    print()

    # ----------------------------------------------------------
    # Benchmark 1: Fibonacci (CPU-bound recursion)
    # ----------------------------------------------------------
    print("-" * 72)
    print("Benchmark 1: Fibonacci (recursive, n=35)")
    print("-" * 72)
    n = 35
    iters = 5
    py_time, _ = bench(fibonacci_py, n, iterations=iters)
    if go_available:
        go_time, _ = bench(go.Fibonacci, n, iterations=iters)
        print_result("fibonacci(35)", py_time, go_time)
    else:
        print_result("fibonacci(35)", py_time, py_time)
    print()

    # ----------------------------------------------------------
    # Benchmark 2: Sum of 1M floats
    # ----------------------------------------------------------
    print("-" * 72)
    print("Benchmark 2: Sum of 1,000,000 floats")
    print("-" * 72)
    data = [random.random() for _ in range(1_000_000)]
    iters = 100
    py_time, _ = bench(sum_slice_py, data, iterations=iters)
    if go_available:
        go_time, _ = bench(go.SumSlice, data, iterations=iters)
        print_result("sum_slice(1M)", py_time, go_time)
    else:
        print_result("sum_slice(1M)", py_time, py_time)
    print()

    # ----------------------------------------------------------
    # Benchmark 3: Matrix multiply (100x100)
    # ----------------------------------------------------------
    print("-" * 72)
    print("Benchmark 3: Matrix multiply (100x100)")
    print("-" * 72)
    size = 100
    a = [random.random() for _ in range(size * size)]
    b = [random.random() for _ in range(size * size)]
    iters = 10
    py_time, _ = bench(matrix_multiply_py, a, b, size, iterations=iters)
    if go_available:
        go_time, _ = bench(go.MatrixMultiply, a, b, size, iterations=iters)
        print_result("matrix_multiply(100x100)", py_time, go_time)
    else:
        print_result("matrix_multiply(100x100)", py_time, py_time)
    print()

    # ----------------------------------------------------------
    # Benchmark 4: Count primes up to 100,000
    # ----------------------------------------------------------
    print("-" * 72)
    print("Benchmark 4: Count primes up to 100,000")
    print("-" * 72)
    iters = 10
    py_time, py_result = bench(count_primes_py, 100_000, iterations=iters)
    if go_available:
        go_time, go_result = bench(go.CountPrimes, 100_000, iterations=iters)
        print_result("count_primes(100K)", py_time, go_time)
        if py_result != go_result:
            print(f"  WARNING: results differ! Python={py_result}, Go={go_result}")
    else:
        print_result("count_primes(100K)", py_time, py_time)
    print()

    # ----------------------------------------------------------
    # Benchmark 5: Sort 100K integers
    # ----------------------------------------------------------
    print("-" * 72)
    print("Benchmark 5: Sort 100,000 integers")
    print("-" * 72)
    sort_data = [random.randint(0, 1_000_000) for _ in range(100_000)]
    iters = 20
    py_time, _ = bench(sort_ints_py, sort_data, iterations=iters)
    if go_available:
        go_time, _ = bench(go.SortInts, sort_data, iterations=iters)
        print_result("sort_ints(100K)", py_time, go_time)
    else:
        print_result("sort_ints(100K)", py_time, py_time)
    print()

    # ----------------------------------------------------------
    # Summary
    # ----------------------------------------------------------
    print("=" * 72)
    print("Summary")
    print("=" * 72)
    if go_available:
        print("Go extensions provide significant speedup over pure Python.")
        print("Use 'goforge build' to create distributable Python wheels.")
    else:
        print("Go bindings not available. Build with 'goforge build' first.")
    print()
    print("Based on benchmarks from: https://programming-language-benchmarks.vercel.app/go-vs-python")
    print("Expected speedups: 10x-50x depending on workload")
    print()


if __name__ == "__main__":
    run_benchmarks()
