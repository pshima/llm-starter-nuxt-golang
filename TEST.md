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