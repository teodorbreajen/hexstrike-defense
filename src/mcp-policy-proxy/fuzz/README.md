# MCP Policy Proxy - Fuzz Testing
## Running Fuzz Tests

### Local Development

Run all fuzz tests with 30 seconds per target:
```bash
cd src/mcp-policy-proxy
make fuzz
```

Run fuzz tests with extended time for deeper coverage:
```bash
make fuzz-dev   # 60 seconds
```

Run a specific fuzz target:
```bash
go test -fuzztime=30s -fuzz=FuzzSanitizeToolInput ./...
go test -fuzztime=30s -fuzz=FuzzIsInternalURL ./...
go test -fuzztime=30s -fuzz=FuzzParseJSONRPC ./...
```

### CI Pipeline

Fuzz tests run in CI with conservative settings:
```bash
make ci-fuzz   # 60 seconds, min-runtime 60s
```

The CI workflow (`test.yml`) includes:
- Unit tests (parallel)
- Integration tests (parallel)
- Fuzz tests (conservative: 60s runtime)

### Corpus Files

Seed corpus files are stored in:
- `testdata/fuzz/FuzzSanitizeToolInput/corpus.json` - Tool input sanitization
- `testdata/fuzz/FuzzIsInternalURL/corpus.txt` - SSRF detection
- `testdata/fuzz/FuzzParseJSONRPC/corpus.json` - JSON-RPC parsing

Go's native fuzzing automatically discovers and saves new inputs to the corpus directory during fuzzing runs.

### Notes

- Fuzz tests require Go 1.18+
- Use `-fuzzminruntime=60s` in CI for consistent results
- Corpus is git-tracked for reproducibility
- New seeds are auto-discovered and saved during fuzz runs