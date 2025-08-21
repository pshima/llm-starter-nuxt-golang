# TEST.md

- Use simple focused tests, avoid large tests that can break easily.
- Test should be quick, completing in milliseconds.
- All tests should have a 5 second timeout maximum
- Playwright should be used for end to end testing and UI/UX needs.  Curl doesn't always work for testing UI and UX issues.  Get Playwright working as one of the first steps.
- Playwright should run with browser UI (headed mode) during development for visibility and debugging
- Keep code coverage above 80%
- We use TDD for this app, write the tests first and then the code to make the tests pass.
- Use docker and test helper scripts whenever possible.  If you need to go around this, come back and fix the helper scripts.
- Use docker-compose.test.yml for test environment configuration and orchestration
- Tests should be easy to run, and easy to debug and fix when broken.
- Golang should use the native golang testing abilities
- Vue/Nuxt should use Vitest for testing
- Always check if testing utilities or code already existing before writing new.
- Always mock external dependencies
- Use miniredis for Redis testing (in-memory Redis mock for Go tests)
- Good tests are fast, isolated, repeatable, self checking, and timely.
- Use a separate agent to write and run tests.

## Repository Testing Patterns

### Redis Testing with Miniredis
```go
// Setup pattern for repository tests
func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
    mr, err := miniredis.Run()
    require.NoError(t, err)
    
    client := redis.NewClient(&redis.Options{
        Addr: mr.Addr(),
    })
    
    t.Cleanup(func() {
        client.Close()
        mr.Close()
    })
    
    return mr, client
}
```

### TTL Testing Strategy
- Use miniredis's `FastForward()` to simulate time passage
- Set TTL on keys and fast-forward to test expiration
- Always verify TTL values after setting them
- Test both before and after expiration states

### Transaction Testing
- Use Redis pipelines for atomic operations
- Test rollback scenarios by simulating failures
- Verify all keys are updated atomically
- Test concurrent access patterns

## Service Testing Patterns

### Mock Generation
```go
// Generate mocks for interfaces
//go:generate mockery --name=UserRepository --output=../mocks --outpkg=mocks
```

### Business Logic Isolation
- Mock all repository dependencies
- Test validation logic separately from persistence
- Use table-driven tests for multiple scenarios
- Test error propagation from repository to service

### Validation Testing Strategy
- Test boundary conditions (min/max lengths)
- Test required field validation
- Test format validation (email, etc.)
- Test business rule enforcement

## Integration Testing Patterns

### Test Server Setup
```go
func SetupTestServer(t *testing.T) *TestServer {
    // Initialize miniredis
    // Wire up all dependencies
    // Create test router
    // Return test server struct with cleanup
}
```

### Session Management in Tests
- Create helper functions for login/logout
- Store cookies between requests
- Test session expiration scenarios
- Verify session isolation between users

### End-to-End Workflow Testing
- Test complete user journeys
- Verify data persistence across requests
- Test error recovery scenarios
- Validate response formats against OpenAPI spec