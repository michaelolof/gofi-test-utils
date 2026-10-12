#!/usr/bin/env python3
"""Compare Gofi request behavior under concurrent HTTP load.

The script starts two pre-built Gofi benchmark servers, alternates samples to
reduce thermal/order bias, and records bombardier latency/throughput alongside
server runtime memory and GC counters. It is intentionally scoped to WS-6; the
general cross-router report remains in httpbench_runner.py.
"""

import argparse
import json
import os
import shutil
import statistics
import subprocess
import sys
import time
from pathlib import Path
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

from httpbench_runner import create_payloads


SCENARIOS = (
    {"name": "small-bind-low", "url": "/json", "payload": "data/small.json", "level": "low"},
    {"name": "small-bind-high", "url": "/json", "payload": "data/small.json", "level": "high"},
    {"name": "small-bind-validate-low", "url": "/json-validate-small", "payload": "data/small.json", "level": "low"},
    {"name": "small-bind-validate-high", "url": "/json-validate-small", "payload": "data/small.json", "level": "high"},
    {"name": "large-bind-low", "url": "/json-large", "payload": "data/large-bind.json", "level": "low"},
    {"name": "large-bind-high", "url": "/json-large", "payload": "data/large-bind.json", "level": "high"},
    {"name": "large-validate-low", "url": "/json-validate-large", "payload": "data/large-validate.json", "level": "low"},
    {"name": "large-validate-high", "url": "/json-validate-large", "payload": "data/large-validate.json", "level": "high"},
    {"name": "large-bind-validate-low", "url": "/json-bind-validate-large", "payload": "data/large-validate.json", "level": "low"},
    {"name": "large-bind-validate-high", "url": "/json-bind-validate-large", "payload": "data/large-validate.json", "level": "high"},
    {"name": "mixed-small-large-low", "level": "low", "mixed": True},
    {"name": "mixed-small-large-high", "level": "high", "mixed": True},
)


def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("--baseline-bin", required=True)
    parser.add_argument("--candidate-bin", required=True)
    parser.add_argument("--output", default="ws6-results.json")
    parser.add_argument("--duration", default="3s")
    parser.add_argument("--samples", type=int, default=3)
    parser.add_argument("--low-concurrency", type=int, default=1)
    parser.add_argument("--high-concurrency", type=int, default=32)
    parser.add_argument("--baseline-port", type=int, default=18080)
    parser.add_argument("--candidate-port", type=int, default=18081)
    parser.add_argument("--retention-cycles", type=int, default=5)
    parser.add_argument(
        "--scenario",
        action="append",
        choices=tuple(scenario["name"] for scenario in SCENARIOS),
        help="run only the named scenario; repeat to select more than one",
    )
    return parser.parse_args()


def wait_until_ready(port, process):
    deadline = time.time() + 15
    while time.time() < deadline:
        if process.poll() is not None:
            raise RuntimeError(f"server on port {port} exited with {process.returncode}")
        try:
            with urlopen(f"http://127.0.0.1:{port}/__ws6/runtime", timeout=1) as response:
                if response.status == 200:
                    return
        except (URLError, TimeoutError):
            time.sleep(0.1)
    raise RuntimeError(f"server on port {port} did not become ready")


def start_server(binary, port):
    environment = os.environ.copy()
    environment["PORT"] = str(port)
    process = subprocess.Popen(
        [str(Path(binary).resolve())],
        env=environment,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.PIPE,
        text=True,
    )
    wait_until_ready(port, process)
    return process


def runtime_stats(port, force_gc=False):
    suffix = "?gc=1" if force_gc else ""
    with urlopen(f"http://127.0.0.1:{port}/__ws6/runtime{suffix}", timeout=10) as response:
        if response.status != 200:
            raise RuntimeError(f"runtime endpoint returned {response.status}")
        return json.load(response)


def preflight(port, url, payload):
    body = Path(payload).read_bytes()
    request = Request(
        f"http://127.0.0.1:{port}{url}",
        data=body,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        response = urlopen(request, timeout=10)
    except HTTPError as error:
        response = error
    response_body = response.read().decode("utf-8")
    status = response.status
    response.close()
    if status != 200 or response_body != "OK":
        raise RuntimeError(f"preflight {url} returned {status} {response_body!r}")


def bombardier_command(binary, port, url, payload, concurrency, duration):
    return [
        binary,
        "--fasthttp",
        "-c", str(concurrency),
        "-d", duration,
        "-l",
        "-m", "POST",
        "-H", "Content-Type: application/json",
        "-f", payload,
        "-o", "json",
        "-p", "r",
        f"http://127.0.0.1:{port}{url}",
    ]


def parse_bombardier(output, expected_status=200):
    decoded = json.loads(output)
    result = decoded["result"]
    expected_key = f"req{expected_status // 100}xx"
    expected_count = result[expected_key]
    total_count = sum(result[key] for key in ("req1xx", "req2xx", "req3xx", "req4xx", "req5xx", "others"))
    if expected_count != total_count or total_count == 0:
        raise RuntimeError(f"unexpected response counts: {result}")
    percentiles = result["latency"]["percentiles"]
    return {
        "requests": total_count,
        "throughputRps": result["rps"]["mean"],
        "meanLatencyUs": result["latency"]["mean"],
        "p50LatencyUs": percentiles["50"],
        "p95LatencyUs": percentiles["95"],
        "p99LatencyUs": percentiles["99"],
    }


def run_single(bombardier, port, scenario, concurrency, duration):
    preflight(port, scenario["url"], scenario["payload"])
    completed = subprocess.run(
        bombardier_command(
            bombardier, port, scenario["url"], scenario["payload"], concurrency, duration
        ),
        check=True,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
    )
    return parse_bombardier(completed.stdout)


def run_mixed(bombardier, port, concurrency, duration):
    streams = (
        {"name": "small", "url": "/json", "payload": "data/small.json"},
        {"name": "large", "url": "/json-large", "payload": "data/large-bind.json"},
    )
    for stream in streams:
        preflight(port, stream["url"], stream["payload"])

    per_stream_concurrency = max(1, concurrency // 2)
    processes = []
    for stream in streams:
        process = subprocess.Popen(
            bombardier_command(
                bombardier, port, stream["url"], stream["payload"], per_stream_concurrency, duration
            ),
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
        )
        processes.append((stream["name"], process))

    results = {}
    for name, process in processes:
        output, _ = process.communicate()
        if process.returncode != 0:
            raise RuntimeError(f"mixed {name} load failed:\n{output}")
        results[name] = parse_bombardier(output)

    # Bombardier does not expose raw latency observations. Keep each stream's
    # percentiles and use the slower stream as a conservative mixed summary.
    return {
        "requests": sum(result["requests"] for result in results.values()),
        "throughputRps": sum(result["throughputRps"] for result in results.values()),
        "meanLatencyUs": max(result["meanLatencyUs"] for result in results.values()),
        "p50LatencyUs": max(result["p50LatencyUs"] for result in results.values()),
        "p95LatencyUs": max(result["p95LatencyUs"] for result in results.values()),
        "p99LatencyUs": max(result["p99LatencyUs"] for result in results.values()),
        "streams": results,
    }


def median_summary(samples):
    keys = ("throughputRps", "meanLatencyUs", "p50LatencyUs", "p95LatencyUs", "p99LatencyUs")
    return {key: statistics.median(sample[key] for sample in samples) for key in keys}


def comparison(baseline, candidate):
    throughput_change = (candidate["throughputRps"] / baseline["throughputRps"] - 1) * 100
    p99_change = (candidate["p99LatencyUs"] / baseline["p99LatencyUs"] - 1) * 100
    return {
        "throughputChangePercent": throughput_change,
        "p99ChangePercent": p99_change,
    }


def retention_check(bombardier, port, concurrency, duration, cycles):
    scenario = {"url": "/json-large", "payload": "data/large-bind.json"}
    observations = []
    runtime_stats(port, force_gc=True)
    for cycle in range(cycles):
        run_single(bombardier, port, scenario, concurrency, duration)
        stats = runtime_stats(port, force_gc=True)
        stats["cycle"] = cycle + 1
        observations.append(stats)

    heap_values = [observation["heapAlloc"] for observation in observations]
    monotonic_growth = all(right > left for left, right in zip(heap_values, heap_values[1:]))
    growth = heap_values[-1] - heap_values[0] if len(heap_values) > 1 else 0
    allowance = max(2 * 1024 * 1024, heap_values[0] // 4)
    return {
        "observations": observations,
        "heapGrowthBytes": growth,
        "allowanceBytes": allowance,
        "monotonicGrowth": monotonic_growth,
        "bounded": not (monotonic_growth and growth > allowance),
    }


def main():
    args = parse_args()
    if args.samples < 1 or args.low_concurrency < 1 or args.high_concurrency < 1:
        raise SystemExit("samples and concurrency values must be positive")

    bombardier = shutil.which("bombardier")
    if bombardier is None:
        raise SystemExit("bombardier is required; install it before running WS-6 validation")

    create_payloads()
    scenarios = tuple(
        scenario
        for scenario in SCENARIOS
        if args.scenario is None or scenario["name"] in args.scenario
    )
    variants = {
        "baseline": {"binary": args.baseline_bin, "port": args.baseline_port},
        "candidate": {"binary": args.candidate_bin, "port": args.candidate_port},
    }
    processes = {}
    results = {
        "configuration": {
            "duration": args.duration,
            "samples": args.samples,
            "lowConcurrency": args.low_concurrency,
            "highConcurrency": args.high_concurrency,
            "bombardier": bombardier,
            "scenarios": [scenario["name"] for scenario in scenarios],
        },
        "variants": {name: {"scenarios": {}} for name in variants},
    }

    try:
        for name, variant in variants.items():
            processes[name] = start_server(variant["binary"], variant["port"])
            results["variants"][name]["runtimeBefore"] = runtime_stats(variant["port"], force_gc=True)

        for scenario in scenarios:
            print(f"{scenario['name']}:", flush=True)
            scenario_samples = {name: [] for name in variants}
            concurrency = args.low_concurrency if scenario["level"] == "low" else args.high_concurrency
            for sample_index in range(args.samples):
                order = ("baseline", "candidate") if sample_index % 2 == 0 else ("candidate", "baseline")
                for name in order:
                    port = variants[name]["port"]
                    if scenario.get("mixed"):
                        sample = run_mixed(bombardier, port, concurrency, args.duration)
                    else:
                        sample = run_single(bombardier, port, scenario, concurrency, args.duration)
                    scenario_samples[name].append(sample)
                    print(
                        f"  {sample_index + 1}/{args.samples} {name}: "
                        f"{sample['throughputRps']:.1f} req/s, p99 {sample['p99LatencyUs']:.0f} us",
                        flush=True,
                    )

            for name in variants:
                results["variants"][name]["scenarios"][scenario["name"]] = {
                    "samples": scenario_samples[name],
                    "median": median_summary(scenario_samples[name]),
                }

        for name, variant in variants.items():
            results["variants"][name]["runtimeAfter"] = runtime_stats(variant["port"])
            results["variants"][name]["retention"] = retention_check(
                bombardier,
                variant["port"],
                args.high_concurrency,
                "1s",
                args.retention_cycles,
            )

        comparisons = {}
        for scenario in scenarios:
            name = scenario["name"]
            baseline = results["variants"]["baseline"]["scenarios"][name]["median"]
            candidate = results["variants"]["candidate"]["scenarios"][name]["median"]
            comparisons[name] = comparison(baseline, candidate)
        results["comparisons"] = comparisons
        selected_small_scenarios = tuple(
            name
            for name in (
                "small-bind-low",
                "small-bind-high",
                "small-bind-validate-low",
                "small-bind-validate-high",
            )
            if name in comparisons
        )
        results["checks"] = {
            "smallThroughputWithinThreePercent": all(
                comparisons[name]["throughputChangePercent"] >= -3
                for name in selected_small_scenarios
            ),
            "p99NotWorse": all(value["p99ChangePercent"] <= 0 for value in comparisons.values()),
            "boundedRetention": all(
                results["variants"][name]["retention"]["bounded"] for name in variants
            ),
        }

        output = Path(args.output)
        output.parent.mkdir(parents=True, exist_ok=True)
        output.write_text(json.dumps(results, indent=2) + "\n", encoding="utf-8")
        print(f"wrote {output}")
        print(json.dumps(results["checks"], indent=2))
    finally:
        for name, process in processes.items():
            process.terminate()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait()
            if process.returncode not in (-15, 0):
                stderr = process.stderr.read() if process.stderr else ""
                print(f"{name} server exited with {process.returncode}: {stderr}", file=sys.stderr)


if __name__ == "__main__":
    main()
