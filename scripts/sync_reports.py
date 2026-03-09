import re
import os
import subprocess

def parse_benchmarks():
    benches = {}
    mems = {}
    go_version = None
    os_info = None
    
    with open('benchmark-results.md', 'r') as f:
        bench_re = re.compile(r'^(Benchmark\w+)-\d+\s+\d+\s+([\d.]+)\s+ns/op\s+(\d+)\s+B/op\s+(\d+)\s+allocs/op')
        mem_re = re.compile(r'^(\w+):\s+(\d+)\s+Bytes')
        section_re = re.compile(r'^#(\w+)\s+Routes:\s+(\d+)')
        
        current_section = ""
        for line in f:
            if line.startswith('goos:'):
                os_info = line.split('goos:')[1].strip()
            if line.startswith('goarch:'):
                current_os = os_info if os_info else ""
                os_info = current_os + "/" + line.split('goarch:')[1].strip()
            if line.startswith('go:'):
                go_version = line.split('go:')[1].strip()
                
            m = section_re.match(line)
            if m:
                current_section = m.group(1)
                continue
            m = mem_re.match(line)
            if m:
                router, bytes_val = m.groups()
                mems[f"{current_section}_{router}"] = int(bytes_val)
                continue
            m = bench_re.match(line)
            if m:
                name, ns, b, allocs = m.groups()
                if name not in benches:
                    benches[name] = []
                benches[name].append((float(ns), int(b), int(allocs)))
                
    avgs = {}
    for k, runs in benches.items():
        ns = sum(r[0] for r in runs) / len(runs)
        b = sum(r[1] for r in runs) / len(runs)
        allocs = sum(r[2] for r in runs) / len(runs)
        avgs[k] = (round(ns), round(b), round(allocs))
        
    return avgs, mems, os_info, go_version

def get_sys_info():
    try:
        cpu = subprocess.check_output("sysctl -n machdep.cpu.brand_string", shell=True).decode().strip()
    except:
        cpu = None
    try:
        ram_bytes = int(subprocess.check_output("sysctl -n hw.memsize", shell=True).decode().strip())
        ram = f"{ram_bytes // (1024**3)} GB"
    except:
        ram = None
    return cpu, ram

def get_http_results():
    http = {}
    endpoints_order = []
    if not os.path.exists('httpbench-results.md'):
        return http, endpoints_order
    with open('httpbench-results.md', 'r') as f:
        for line in f:
            if line.startswith('| ') and 'Endpoint' not in line and '---' not in line:
                parts = [p.strip() for p in line.split('|')]
                if len(parts) >= 8:
                    endpoint = parts[1].replace('`', '').strip()
                    method_url = parts[2].replace('`', '').strip()
                    try:
                        http[endpoint] = {
                            'method_url': method_url,
                            'Gofi': float(parts[3].replace('**', '').replace(',', '').strip()),
                            'Chi': float(parts[4].replace('**', '').replace(',', '').strip()),
                            'Echo': float(parts[5].replace('**', '').replace(',', '').strip()),
                            'Gin': float(parts[6].replace('**', '').replace(',', '').strip()),
                            'Fiber': float(parts[7].replace('**', '').replace(',', '').strip()),
                        }
                        endpoints_order.append(endpoint)
                    except ValueError:
                        pass
    return http, endpoints_order

def fmt_num(n):
    s = str(int(n))
    if len(s) <= 3: return s
    return ",".join([s[max(0, i-3):i] for i in range(len(s), 0, -3)][::-1])

def process_file(filepath, avgs, mems, http, os_info, go_version, cpu, ram):
    if not os.path.exists(filepath): return
    with open(filepath, 'r') as f:
        lines = f.readlines()
        
    out = []
    i = 0
    winner_str = ""
    winner_val = 0
    winner_b = 0
    winner_allocs = 0
    in_key_takeaways = False
    in_schema_overhead_verdict = False

    while i < len(lines):
        line = lines[i]
        
        # Stop completely if we hit Key Takeaways
        if line.startswith("## Key Takeaways"):
            in_key_takeaways = True
        if in_key_takeaways:
            i += 1
            continue

        # Rewrite Test Environment
        if line.startswith("## Test Environment"):
            out.append(line)
            while i+1 < len(lines) and (lines[i+1].startswith("-") or lines[i+1].strip() == ""):
                i += 1
            if cpu: out.append(f"- **CPU:** {cpu}\n")
            if ram: out.append(f"- **RAM:** {ram}\n")
            if os_info: out.append(f"- **OS:** macOS ({os_info})\n")
            if go_version: out.append(f"- **Go Version:** {go_version.replace('go', '')}\n")
            i += 1
            continue
            
        # Clean Schema Overhead verbose outputs
        if line.startswith("## Schema Overhead"):
            out.append(line)
            while i+1 < len(lines) and not lines[i+1].startswith("---") and not lines[i+1].startswith("##"):
                i += 1
                next_line = lines[i]
                if next_line.startswith("> "):
                    continue # Skip the static explanation quote
                out.append(next_line)
            i += 1
            continue

        out.append(line)
        
        # Micro Benchmarks & JSON Tables
        if line.startswith("### ") or line.startswith("#### "):
            header_text = line.strip("# \n").split("—")[0].strip()
            suffix_map = {
                "Static Route": "Static",
                "Single Param": "Param",
                "5 Params": "Param5",
                "20 Params": "Param20",
                "Param Write": "ParamWrite",
                "Multi Param": "MultiParam",
                "Wildcard": "Wildcard",
                "Deep Nesting": "Deep",
                "404 Handling": "404",
                "JSON Binding (Small Payload)": "BindJSON_Small",
                "JSON Response (100 items)": "JSONResponse_Large",
                "Concurrency (Parallel Requests)": "Parallel",
                "Route Groups": "RouteGroup"
            }
            suffix = suffix_map.get(header_text)
            
            if suffix and i + 2 < len(lines) and "| Router |" in lines[i+1]:
                i += 1
                out.append(lines[i]) # header
                i += 1
                out.append(lines[i]) # separator
                
                rows = []
                while i + 1 < len(lines) and lines[i+1].startswith("|"):
                    i += 1
                    r_line = lines[i]
                    if "---" in r_line: continue
                    parts = [p.strip() for p in r_line.split("|")]
                    if len(parts) >= 5:
                        router_display = parts[1].replace("**", "")
                        prefix = router_display.replace(" + Schema", "S").replace(" ", "")
                        key = f"Benchmark{prefix}_{suffix}"
                        if key in avgs:
                            ns, b, allocs = avgs[key]
                            rows.append({"display": router_display, "prefix": prefix, "ns": ns, "b": b, "allocs": allocs})
                        else:
                            ns = int(parts[2].replace("**", "").replace(",", "")) if parts[2].replace("**", "").replace(",", "").isdigit() else 0
                            b = int(parts[3].replace("**", "").replace(",", "")) if parts[3].replace("**", "").replace(",", "").isdigit() else 0
                            allocs = int(parts[4].replace("**", "").replace(",", "")) if parts[4].replace("**", "").replace(",", "").isdigit() else 0
                            rows.append({"display": router_display, "prefix": prefix, "ns": ns, "b": b, "allocs": allocs})

                if rows:
                    min_ns = min((r["ns"] for r in rows if r["ns"] > 0), default=0)
                    for r in rows:
                        is_winner = r["ns"] == min_ns and min_ns > 0
                        if is_winner:
                            winner_str = r['display']
                            winner_val = r['ns']
                            winner_b = r['b']
                            winner_allocs = r['allocs']
                        w = "**" if is_winner else ""
                        out.append(f"| {w}{r['display']}{w} | {w}{fmt_num(r['ns'])}{w} | {w}{fmt_num(r['b'])}{w} | {w}{fmt_num(r['allocs'])}{w} |\n")
        
        # Replace the winner text entirely
        if line.startswith("> 🥇 **"):
            if winner_str:
                out[-1] = f"> 🥇 **{winner_str}** — {fmt_num(winner_val)} ns.\n"
                winner_str = ""
                
        i += 1
        
    with open(filepath, 'w') as f:
        f.writelines(out)

def sync_readme(filepath, http, endpoints_order, os_info, go_version, cpu, ram):
    if not os.path.exists(filepath): return
    with open(filepath, 'r') as f:
        lines = f.readlines()
        
    out = []
    i = 0
    while i < len(lines):
        line = lines[i]
        
        # Rewrite Test Environment
        if line.startswith("## Test Environment"):
            out.append(line)
            while i+1 < len(lines) and (lines[i+1].startswith("-") or "All benchmarks" in lines[i+1] or lines[i+1].strip() == ""):
                if "All benchmarks" in lines[i+1]:
                    i += 1
                    out.append(lines[i])
                else:
                    i += 1
            if cpu: out.append(f"- **CPU:** {cpu}\n")
            if ram: out.append(f"- **RAM:** {ram}\n")
            if os_info: out.append(f"- **OS:** macOS ({os_info})\n")
            if go_version: out.append(f"- **Go Version:** {go_version.replace('go', '')}\n")
            i += 1
            continue

        out.append(line)
        if "### Results Overview" in line and i + 3 < len(lines) and "Endpoint" in lines[i+2]:
            i += 1
            out.append(lines[i]) # empty
            i += 1
            out.append("| Case | Endpoint | Gofi (fasthttp) | Fiber (fasthttp) | Chi (net/http) | Gin (net/http) | Echo (net/http) | Winner |\n")
            i += 1
            out.append("|---|---|---|---|---|---|---|---|\n")
            
            # Skip old table rows
            while i + 1 < len(lines) and lines[i+1].startswith("|"):
                i += 1
                
            # Drop the paragraph directly under the table if it's static
            if i + 1 < len(lines) and lines[i+1].startswith("**What this means:**"):
                i += 1
                
            for raw_endpoint in endpoints_order:
                if raw_endpoint in http:
                    perf = http[raw_endpoint]
                    method_url = perf['method_url']
                    g = perf['Gofi']; fb = perf['Fiber']; c = perf['Chi']; gn = perf['Gin']; e = perf['Echo']
                    best = max(g, fb, c, gn, e)
                    def w(v, name):
                        return f"**{fmt_num(v)}**" if v == best else fmt_num(v)
                    winner = "🏆 **Gofi**" if best == g else ("🏆 **Fiber**" if best == fb else "🏆 **Chi**" if best == c else "🏆 **Gin**" if best == gn else "🏆 **Echo**")
                    out.append(f"| {raw_endpoint} | `{method_url}` | {w(g, 'Gofi')} | {w(fb, 'Fiber')} | {w(c, 'Chi')} | {w(gn, 'Gin')} | {w(e, 'Echo')} | {winner} |\n")
        i += 1
        
    with open(filepath, 'w') as f:
        f.writelines(out)

def main():
    avgs, mems, os_info, go_version = parse_benchmarks()
    cpu, ram = get_sys_info()
    http, endpoints_order = get_http_results()
    
    for md in ["gofi_vs_chi_benchmarks.md", "gofi_vs_echo_benchmarks.md", "gofi_vs_gin_benchmarks.md", "gofi_vs_fibre_benchmarks.md"]:
        process_file(md, avgs, mems, http, os_info, go_version, cpu, ram)
        print(f"Synced {md}")
        
    sync_readme("README.md", http, endpoints_order, os_info, go_version, cpu, ram)
    print("Synced README.md")

if __name__ == "__main__":
    main()
