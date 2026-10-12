import copy
import json
import os
import re
import shutil
import subprocess
import time
import uuid
from datetime import datetime, timezone
from urllib.error import HTTPError
from urllib.request import Request, urlopen


ROUTERS = ["gofi", "chi", "echo", "gin", "fiber"]
PORTS = ["8080", "8081", "8082", "8083", "8084"]
DURATION = "5s"
CONNECTIONS = 125


def create_payloads():
    small = {"id": 1, "name": "test"}
    now_iso = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")

    large_bind = []
    large_validate = []
    for i in range(50):
        large_bind.append({
            "id": f"uuid-{i}",
            "userId": i,
            "accountType": "premium",
            "isVerified": i % 2 == 0,
            "balance": 1500.50 + i,
            "createdAt": now_iso,
            "metadata": {"source": "web", "version": "1.0"},
            "events": [
                {
                    "eventId": f"evt-{i}-1",
                    "timestamp": now_iso,
                    "tags": ["login", "success"],
                    "metrics": {"cpuUsage": 45.2, "memoryBytes": 1024000, "active": True},
                },
                {
                    "eventId": f"evt-{i}-2",
                    "timestamp": now_iso,
                    "tags": ["view", "page"],
                    "metrics": {"cpuUsage": 55.1, "memoryBytes": 2048000, "active": False},
                },
            ],
        })

        large_validate.append({
            "id": str(uuid.uuid4()),
            "userId": i + 1,
            "accountType": "premium",
            "isVerified": i % 2 == 0,
            "balance": 1500.50 + i,
            "createdAt": now_iso,
            "metadata": {"source": "web", "version": "1.0"},
            "events": [
                {
                    "eventId": f"evt{i}1",
                    "timestamp": now_iso,
                    "tags": ["login", "success", "choir"],
                    "metrics": {"cpuUsage": 45.2, "memoryBytes": 1024000, "active": True},
                },
                {
                    "eventId": f"evt{i}2",
                    "timestamp": now_iso,
                    "tags": ["view", "page", "stuff"],
                    "metrics": {"cpuUsage": 55.1, "memoryBytes": 2048000, "active": False},
                },
            ],
        })

    early_failure = copy.deepcopy(large_validate)
    early_failure[0]["id"] = "invalid-uuid"
    late_failure = copy.deepcopy(large_validate)
    late_failure[-1]["events"][-1]["metrics"]["cpuUsage"] = 101

    os.makedirs("data", exist_ok=True)
    payloads = {
        "small.json": small,
        "large-bind.json": large_bind,
        "large-validate.json": large_validate,
        "large-fail-early.json": early_failure,
        "large-fail-late.json": late_failure,
    }
    for filename, payload in payloads.items():
        with open(os.path.join("data", filename), "w", encoding="utf-8") as payload_file:
            json.dump(payload, payload_file)

    with open("data/formdata.txt", "w", encoding="utf-8") as form_file:
        form_file.write("id=1&name=test")

    multipart_body = (
        "--myboundary\r\n"
        "Content-Disposition: form-data; name=\"id\"\r\n\r\n"
        "1\r\n"
        "--myboundary\r\n"
        "Content-Disposition: form-data; name=\"name\"\r\n\r\n"
        "test\r\n"
        "--myboundary--\r\n"
    )
    with open("data/multipart.txt", "w", encoding="utf-8") as multipart_file:
        multipart_file.write(multipart_body)


def scenario(name, method, url, content_type="", payload_file="", expected_status=200,
             expected_body=None, expected_json_length=None, expected_json=None,
             expected_body_contains=None):
    return {
        "name": name,
        "method": method,
        "url": url,
        "content_type": content_type,
        "payload_file": payload_file,
        "expected_status": expected_status,
        "expected_body": expected_body,
        "expected_json_length": expected_json_length,
        "expected_json": expected_json,
        "expected_body_contains": expected_body_contains,
    }


ENDPOINTS = [
    scenario("Static Route", "GET", "/", expected_body="OK"),
    scenario("Single Param", "GET", "/user/gordon", expected_body="gordon"),
    scenario("Multi Param", "GET", "/users/123/posts/456", expected_body="123:456"),
    scenario("Middleware Chain", "GET", "/middlewares", expected_body="OK"),
    scenario("Query Processing", "GET", "/query?q=searchterm&limit=10", expected_body="searchterm10"),
    scenario("JSON Bind (Small)", "POST", "/json", "application/json", "./data/small.json", expected_body="OK"),
    scenario("JSON Response (Small)", "GET", "/json-response", expected_json_length=100),
    scenario("JSON Bind Only (Large)", "POST", "/json-large", "application/json", "./data/large-bind.json", expected_body="OK"),
    scenario("JSON Response (Large)", "GET", "/json-response-large", expected_json_length=50),
    scenario("JSON Bind + Validate (Small)", "POST", "/json-validate-small", "application/json", "./data/small.json", expected_body="OK"),
    scenario("JSON Validate Response (Small)", "GET", "/json-response-validate-small", expected_json={"id": 1, "name": "test"}),
    scenario("JSON Validate Only (Large)", "POST", "/json-validate-large", "application/json", "./data/large-validate.json", expected_body="OK"),
    scenario("JSON Bind + Validate (Large)", "POST", "/json-bind-validate-large", "application/json", "./data/large-validate.json", expected_body="OK"),
    scenario("JSON Validation Failure (Early)", "POST", "/json-bind-validate-large", "application/json", "./data/large-fail-early.json", 400, expected_body_contains="bad"),
    scenario("JSON Validation Failure (Late/Deep)", "POST", "/json-bind-validate-large", "application/json", "./data/large-fail-late.json", 400, expected_body_contains="bad"),
    scenario("JSON Validate Response (Large)", "GET", "/json-response-validate-large", expected_json_length=50),
    scenario("Multipart Bind", "POST", "/multipart", "multipart/form-data; boundary=myboundary", "./data/multipart.txt", expected_body="OK"),
    scenario("FormData Bind", "POST", "/formdata", "application/x-www-form-urlencoded", "./data/formdata.txt", expected_body="OK"),
]


def run_cmd(command):
    completed = subprocess.run(command, check=True, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    return completed.stdout


def preflight(port, endpoint):
    payload = None
    if endpoint["payload_file"]:
        with open(endpoint["payload_file"], "rb") as payload_file:
            payload = payload_file.read()

    headers = {}
    if endpoint["content_type"]:
        headers["Content-Type"] = endpoint["content_type"]

    request = Request(
        f"http://localhost:{port}{endpoint['url']}",
        data=payload,
        headers=headers,
        method=endpoint["method"],
    )
    try:
        response = urlopen(request, timeout=10)
    except HTTPError as error:
        response = error

    status = response.status
    body = response.read()
    response.close()
    if status != endpoint["expected_status"]:
        raise RuntimeError(
            f"{endpoint['name']} preflight returned status {status}, "
            f"expected {endpoint['expected_status']}; body={body!r}"
        )

    text_body = body.decode("utf-8").strip()
    if endpoint["expected_body"] is not None and text_body != endpoint["expected_body"]:
        raise RuntimeError(
            f"{endpoint['name']} preflight returned body {text_body!r}, "
            f"expected {endpoint['expected_body']!r}"
        )
    if endpoint["expected_body_contains"] is not None and endpoint["expected_body_contains"].casefold() not in text_body.casefold():
        raise RuntimeError(
            f"{endpoint['name']} preflight body {text_body!r} does not contain "
            f"{endpoint['expected_body_contains']!r}"
        )
    if endpoint["expected_json"] is not None or endpoint["expected_json_length"] is not None:
        try:
            decoded = json.loads(body)
        except json.JSONDecodeError as error:
            raise RuntimeError(f"{endpoint['name']} preflight returned invalid JSON") from error
        if endpoint["expected_json"] is not None and decoded != endpoint["expected_json"]:
            raise RuntimeError(
                f"{endpoint['name']} preflight returned JSON {decoded!r}, "
                f"expected {endpoint['expected_json']!r}"
            )
        if endpoint["expected_json_length"] is not None and len(decoded) != endpoint["expected_json_length"]:
            raise RuntimeError(
                f"{endpoint['name']} preflight returned {len(decoded)} JSON items, "
                f"expected {endpoint['expected_json_length']}"
            )


def main():
    print("Creating payloads...")
    create_payloads()

    bombardier = shutil.which("bombardier")
    if bombardier is None:
        print("Installing bombardier...")
        run_cmd(["go", "install", "github.com/codesenberg/bombardier@latest"])
        go_path = run_cmd(["go", "env", "GOPATH"]).strip()
        bombardier = os.path.join(go_path, "bin", "bombardier")

    print("\nBuilding binaries...")
    os.makedirs("bin", exist_ok=True)
    for router in ROUTERS:
        print(f"  Building {router}...")
        run_cmd(["go", "build", "-o", f"bin/{router}", f"./cmd/httpbench/{router}/"])

    print(f"\n{'=' * 44}")
    print(f" Starting HTTP Benchmarks ({DURATION}, {CONNECTIONS} conns)")
    print(f"{'=' * 44}\n")

    results = {}
    for index, router in enumerate(ROUTERS):
        port = PORTS[index]
        print(f"Testing {router} on port {port}...")
        environment = os.environ.copy()
        environment["PORT"] = port
        process = subprocess.Popen(
            [f"./bin/{router}"],
            env=environment,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
        time.sleep(2)
        results[router] = {}

        try:
            for endpoint in ENDPOINTS:
                print(f"  -> {endpoint['name']} ({endpoint['method']} {endpoint['url']})")
                preflight(port, endpoint)

                command = [
                    bombardier,
                    "-c", str(CONNECTIONS),
                    "-d", DURATION,
                    "-m", endpoint["method"],
                ]
                if endpoint["content_type"]:
                    command.extend(["-H", f"Content-Type: {endpoint['content_type']}"])
                if endpoint["payload_file"]:
                    command.extend(["-f", endpoint["payload_file"]])
                command.append(f"http://localhost:{port}{endpoint['url']}")

                output = run_cmd(command)
                match = re.search(r"Reqs/sec\s+([0-9.,]+)", output)
                if match is None:
                    raise RuntimeError(f"could not find Reqs/sec in bombardier output for {endpoint['name']}")
                results[router][endpoint["name"]] = match.group(1).strip()
        finally:
            process.terminate()
            process.wait()
            time.sleep(1)

    print("\nFormatting results...")
    results_file = "httpbench-results.md"
    with open(results_file, "w", encoding="utf-8") as output_file:
        output_file.write("# HTTP Load Test Results\n\n")
        output_file.write(f"Tests run with `bombardier` for {DURATION} with {CONNECTIONS} concurrent connections.\n\n")
        output_file.write("| Endpoint | Method | " + " | ".join(router.capitalize() for router in ROUTERS) + " | Winner |\n")
        output_file.write("|---" * (3 + len(ROUTERS)) + "|\n")

        for endpoint in ENDPOINTS:
            row = [endpoint["name"], f"`{endpoint['method']} {endpoint['url']}`"]
            values = []
            winner = ""
            best_value = -1
            for router in ROUTERS:
                value_string = results[router][endpoint["name"]]
                value = float(value_string.replace(",", ""))
                values.append((router, value, value_string))
                if value > best_value:
                    best_value = value
                    winner = router

            for router, _, value_string in values:
                row.append(f"**{value_string}**" if router == winner else value_string)
            row.append(f"**{winner.capitalize()}**")
            output_file.write("| " + " | ".join(row) + " |\n")

    print(f"Done! Results written to {results_file}")
    with open(results_file, "r", encoding="utf-8") as output_file:
        print(output_file.read())


if __name__ == "__main__":
    main()
