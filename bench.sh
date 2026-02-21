#!/usr/bin/env bash
# bench.sh - run all benchmarks and dump results into README.md
#
# Usage:
#   ./bench.sh        # run benchmarks and update README
#
# The script records system info, runs `go test -bench` and then
# splices the output into README.md between the markers
# <!-- BENCH-START --> and <!-- BENCH-END -->.  If the markers
# are not present it will append the results at the end.

set -euo pipefail

# collect system information
sysinfo=$(cat <<'EOF'
$(go env GOOS)
$(go env GOARCH)
$(sysctl -n machdep.cpu.brand_string 2>/dev/null || echo "cpu: unknown")
$(go version | awk '{print $3}')
EOF
)

# run the benchmarks
bench=$(go test ./internal/benchmarks/ -bench=. -benchmem -count=3 2>&1)

# build markdown snippet
cat <<EOF > /tmp/bench-results.md
## System Info

\\`\\\`\\\`
goos: $(echo "$sysinfo" | head -1)
goarch: $(echo "$sysinfo" | sed -n '2p')
cpu: $(echo "$sysinfo" | sed -n '3p')
go: $(echo "$sysinfo" | sed -n '4p')
benchmarks: -bench=. -benchmem -count=3
\\`\\\`\\\`

\\`\\\`
$bench
\\`\\\`
EOF

# inject into README.md between markers, create markers if missing
if grep -q "<!-- BENCH-START -->" README.md; then
  awk '/<!-- BENCH-START -->/{print; system("cat /tmp/bench-results.md"); skip=1; next}
       /<!-- BENCH-END -->/{skip=0}
       !skip {print}' README.md > README.tmp && mv README.tmp README.md
else
  printf "\n<!-- BENCH-START -->\n" >> README.md
  cat /tmp/bench-results.md >> README.md
  printf "\n<!-- BENCH-END -->\n" >> README.md
fi

echo "Benchmarks run and README.md updated."
