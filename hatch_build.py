"""Hatch build hook: compile the Go binary before packaging."""

import os
import platform
import subprocess
import sys

from hatchling.builders.hooks.plugin.interface import BuildHookInterface


def _get_goos():
    system = platform.system().lower()
    return {"darwin": "darwin", "linux": "linux", "windows": "windows"}.get(system, system)


def _get_goarch():
    machine = platform.machine().lower()
    if machine in ("x86_64", "amd64"):
        return "amd64"
    elif machine in ("arm64", "aarch64"):
        return "arm64"
    return machine


class CustomBuildHook(BuildHookInterface):
    def initialize(self, version, build_data):
        root = os.path.dirname(os.path.abspath(__file__))
        go_dir = os.path.join(root, "goforge")
        bin_dir = os.path.join(root, "goforge_pkg", "bin")
        os.makedirs(bin_dir, exist_ok=True)

        binary_name = "goforge"
        goos = _get_goos()
        if goos == "windows":
            binary_name += ".exe"
        output = os.path.join(bin_dir, binary_name)

        env = os.environ.copy()
        env["GOOS"] = goos
        env["GOARCH"] = _get_goarch()
        env["CGO_ENABLED"] = "0"

        cmd = [
            "go",
            "build",
            "-ldflags",
            "-s -w",
            "-o",
            output,
            ".",
        ]

        print(f"GoForge: Compiling Go binary for {goos}/{env['GOARCH']}...")
        result = subprocess.run(cmd, cwd=go_dir, env=env, capture_output=True, text=True)

        if result.returncode != 0:
            print(f"GoForge: Go compilation failed:\n{result.stderr}", file=sys.stderr)
            raise RuntimeError(f"Go compilation failed for {goos}/{env['GOARCH']}")

        size_mb = os.path.getsize(output) / (1024 * 1024)
        print(f"GoForge: Built goforge binary ({size_mb:.1f} MB)")
