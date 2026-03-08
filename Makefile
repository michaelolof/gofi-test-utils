.PHONY: all bench httpbench sync clean

all: bench httpbench sync

bench:
	@echo "Running Routing Microbenchmarks..."
	@go run ./cmd/report/

httpbench:
	@echo "Running HTTP Load Tests..."
	@python3 ./scripts/httpbench_runner.py

sync:
	@echo "Syncing benchmark reports..."
	@python3 ./scripts/sync_reports.py

clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@rm -f benchmark-results.md httpbench-results.md
