# 🧪 Integration and Unit Tests

## 📌 Why We Need Integration and Unit Tests

Automated testing ensures that our application behaves as expected and remains stable over time.

- **Unit Tests**: Verify individual functions or methods in isolation. Fast and precise.
- **Integration Tests**: Verify how multiple components interact. Useful for catching edge cases in system communication.

**Benefits:**
- Catch bugs early
- Prevent regressions
- Improve confidence in code changes
- Enable continuous integration and delivery

---

## ✍️ How to Write Tests

### Unit Tests

- Should cover a single function or method
- Should not rely on external systems (use mocks or stubs)
- Fast to run and easy to debug

### Integration Tests

- Validate the interaction between components (e.g., APIs)
- Can use real services, containers, or test environments
- Slower than unit tests, but critical for system confidence

## 📂 Test Structure

```
├── test/                      # Integration test suite
│   ├── main_test.go
│   └── user_test.go
├── testdata/                  # Test fixtures and sample payloads
│   └── mock_user.json
├── testUtils/                 # Shared test utils
│   └── env.go                 # Handles env loading and fallback
├── go.mod
```

- Unit tests live next to implementation code
- Integration tests live under `/test`
- Shared test resources go into `/testdata`
- Environment-related helpers live in `/testUtils`

---

## 🚀 How to Run the Tests

### Run All Integration Tests by Network and Tag
```bash
make integration-test TAGS=e2e NETWORK=mainnet TEST=./test
```
### Run a Specific Test
```bash
ENV_FILE=test/testEnv/.env.mainnet go test -v -run TestHealthzEndpoint ./test/healthz_test.go
```

### Copy Example Env File for Local Testing
If the file `test/testEnv/.env.mainnet` is missing, it will be copied automatically from `.env.example` on first run.
You can also copy it manually:
```bash
cp .env.example test/testEnv/.env.mainnet
```
Then update the values inside it as needed.

---

## ✅ Best Practices

- **Keep all example variables in `.env.example`** for clarity
- **Don't commit real `.env` files**, use `.gitignore`
- **Use matrix testing in GitHub Actions** to run tests across multiple environments automatically
- **Use `testUtils/env.go`** to fallback to `.env.example` if real env file is missing (for local usage)

---

## ✅ CI/CD Notes

In GitHub Actions:
- The matrix runner sets the `ENV_FILE` path based on the environment
- The test suite picks the correct `.env` file automatically
- No copying is done in CI – it expects the file already to exist in the repo under `test/testEnv/.env.<network>`

If secrets are needed, pass them through `secrets:` and assign them via `echo "VAR=value" >> $GITHUB_ENV` in the workflow
