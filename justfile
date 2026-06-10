# dado — build & performance tooling
#
# Benchmark workflow:
#   just bench-baseline   # capture current numbers into bench/baseline.txt (commit it)
#   ...make a change...
#   just bench-new        # capture into bench/new.txt
#   just bench-compare    # benchstat baseline.txt new.txt
#   just bench-profile    # write cpu/mem profiles for the core package
#
# benchstat:  go install golang.org/x/perf/cmd/benchstat@latest

bench_pkgs  := "./core ./components"
bench_flags := "-run=^$ -bench=. -benchmem -benchtime=200x -count=6"

# list recipes
default:
    @just --list

test:
    go test ./...

# Quick one-shot run, printed to stdout.
bench:
    go test -run=^$ -bench=. -benchmem {{bench_pkgs}}

# Core package only — fastest feedback loop while optimizing the render path.
bench-core:
    go test -run=^$ -bench=. -benchmem -benchtime=200x ./core

# Capture a comparable baseline (commit bench/baseline.txt).
bench-baseline:
    mkdir -p bench
    go test {{bench_flags}} {{bench_pkgs}} | tee bench/baseline.txt

bench-new:
    mkdir -p bench
    go test {{bench_flags}} {{bench_pkgs}} | tee bench/new.txt

# Requires benchstat (see header). -count=6 in the flags gives it variance.
bench-compare:
    benchstat bench/baseline.txt bench/new.txt

# CPU + allocation profiles for the core render path. Inspect with:
#   go tool pprof -http=: bench/core.cpu.prof
#   go tool pprof -http=: bench/core.mem.prof
bench-profile:
    mkdir -p bench
    go test -run=^$ -bench=. -benchmem -benchtime=500x \
        -cpuprofile=bench/core.cpu.prof \
        -memprofile=bench/core.mem.prof ./core
