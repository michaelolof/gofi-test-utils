import subprocess
import time
import re
import os
import signal
import json
from datetime import datetime, timezone

ROUTERS = ["gofi", "chi", "echo", "gin", "fiber"]
PORTS = ["8080", "8081", "8082", "8083", "8084"]
DURATION = "5s"
CONNECTIONS = 125

def create_payloads():
    small = {"id": 1, "name": "test"}
    
    # Generate 50 complex payloads for realistically large benchmarking
    large = []
    now_iso = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    for i in range(50):
        large.append({
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
                    "metrics": {"cpuUsage": 45.2, "memoryBytes": 1024000, "active": True}
                },
                {
                    "eventId": f"evt-{i}-2",
                    "timestamp": now_iso,
                    "tags": ["view", "page"],
                    "metrics": {"cpuUsage": 55.1, "memoryBytes": 2048000, "active": False}
                }
            ]
        })
    
    os.makedirs("data", exist_ok=True)
    with open("data/small.json", "w") as f: json.dump(small, f)
    with open("data/large.json", "w") as f: json.dump(large, f)
    
    with open("data/formdata.txt", "w") as f: f.write("id=1&name=test")
    
    multipart_body = (
        "--myboundary\r\n"
        "Content-Disposition: form-data; name=\"id\"\r\n\r\n"
        "1\r\n"
        "--myboundary\r\n"
        "Content-Disposition: form-data; name=\"name\"\r\n\r\n"
        "test\r\n"
        "--myboundary--\r\n"
    )
    with open("data/multipart.txt", "w") as f: f.write(multipart_body)

ENDPOINTS = [
    ("Static Route", "GET", "/", "", ""),
    ("Single Param", "GET", "/user/gordon", "", ""),
    ("Multi Param", "GET", "/users/123/posts/456", "", ""),
    ("Middleware Chain", "GET", "/middlewares", "", ""),
    ("Query Processing", "GET", "/query?q=searchterm&limit=10", "", ""),
    ("JSON Bind (Small)", "POST", "/json", "application/json", "./data/small.json"),
    ("JSON Response (Small)", "GET", "/json-response", "", ""),
    ("JSON Bind (Large)", "POST", "/json-large", "application/json", "./data/large.json"),
    ("JSON Response (Large)", "GET", "/json-response-large", "", ""),
    ("JSON Validate (Small)", "POST", "/json-validate-small", "application/json", "./data/small.json"),
    ("JSON Validate Response (Small)", "GET", "/json-response-validate-small", "", ""),
    ("JSON Validate (Large)", "POST", "/json-validate-large", "application/json", "./data/large.json"),
    ("JSON Validate Response (Large)", "GET", "/json-response-validate-large", "", ""),
    ("Multipart Bind", "POST", "/multipart", "multipart/form-data; boundary=myboundary", "./data/multipart.txt"),
    ("FormData Bind", "POST", "/formdata", "application/x-www-form-urlencoded", "./data/formdata.txt"),
]

def run_cmd(cmd):
    return subprocess.check_output(cmd, shell=True).decode('utf-8')

def main():
    print("Creating payloads...")
    create_payloads()
    
    print("Checking for bombardier...")
    try:
        run_cmd("bombardier --help")
    except subprocess.CalledProcessError:
        print("Installing bombardier...")
        run_cmd("go install github.com/codesenberg/bombardier@latest")
        if "GOPATH" in os.environ:
            os.environ["PATH"] += os.pathsep + os.path.join(os.environ["GOPATH"], "bin")
        else:
            os.environ["PATH"] += os.pathsep + os.path.join(os.environ["HOME"], "go", "bin")

    print("\nBuilding binaries...")
    run_cmd("mkdir -p bin")
    for r in ROUTERS:
        print(f"  Building {r}...")
        run_cmd(f"go build -o bin/{r} ./cmd/httpbench/{r}/")

    print(f"\n============================================")
    print(f" Starting HTTP Benchmarks ({DURATION}, {CONNECTIONS} conns)")
    print(f"============================================\n")

    results = {}
    
    for i, router in enumerate(ROUTERS):
        port = PORTS[i]
        print(f"Testing {router} on port {port}...")
        
        env = os.environ.copy()
        env["PORT"] = port
        
        proc = subprocess.Popen([f"./bin/{router}"], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        time.sleep(2) # wait for startup
        
        results[router] = {}
        
        for name, method, url, ctype, pfile in ENDPOINTS:
            print(f"  -> {name} ({method} {url})")
            full_url = f"http://localhost:{port}{url}"
            
            cmd = f"bombardier -c {CONNECTIONS} -d {DURATION} -m {method}"
            if ctype:
                cmd += f" -H 'Content-Type: {ctype}'"
            if pfile:
                cmd += f" -f {pfile}"
            cmd += f" \"{full_url}\""
                
            out = run_cmd(cmd)
            
            # Extract Reqs/sec
            match = re.search(r"Reqs/sec\s+([0-9.,]+)", out)
            if match:
                reqs = match.group(1).strip()
                results[router][name] = reqs
            else:
                results[router][name] = "0"
                
        # cleanup
        proc.terminate()
        proc.wait()
        time.sleep(1)

    print("\nFormatting results...")
    md_file = "httpbench-results.md"
    with open(md_file, "w") as f:
        f.write("# HTTP Load Test Results\n\n")
        f.write(f"Tests run with `bombardier` for {DURATION} with {CONNECTIONS} concurrent connections.\n\n")
        f.write("| Endpoint | Method | " + " | ".join(r.capitalize() for r in ROUTERS) + " | Winner |\n")
        f.write("|---" * (3 + len(ROUTERS)) + "|\n")
        
        for name, method, url, _, _ in ENDPOINTS:
            row = [name, f"`{method} {url}`"]
            
            best_val = -1
            winner = ""
            
            # Get values
            vals = []
            for r in ROUTERS:
                val_str = results[r][name]
                clean_val = float(val_str.replace(',', ''))
                vals.append((r, clean_val, val_str))
                
                if clean_val > best_val:
                    best_val = clean_val
                    winner = r
                    
            for r, clean, string in vals:
                if r == winner:
                    row.append(f"**{string}**")
                else:
                    row.append(string)
                    
            row.append(f"**{winner.capitalize()}**")
            f.write("| " + " | ".join(row) + " |\n")

    print(f"Done! Results written to {md_file}")
    with open(md_file, "r") as f:
        print(f.read())

if __name__ == "__main__":
    main()
