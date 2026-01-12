# Unit tests instructions

## Libraries
1. Use the `github.com/stretchr/testify` library, particularly the `assert` and `require` packages, for writing assertions.
2. Use the `github.com/reugn/go-quartz` library for time mocking.
3. Use the `github.com/spf13/afero` and `afero.NewMemMapFs` for mocking file system.

## Planning
4. List the scenarios you plan to cover and confirm them with the user before writing tests.
5. Target one SUT (function/method) per test suite. If validating behavior that spans multiple units, ask the user for confirmation.

## Code
6. Write one behavior per test.
7. Name tests using Test<Thing>_<When>_<Then>. Use descriptive t.Run subtest names too.
8. Keep strict flow: Arrange - Act - Assert. 
9. Don’t do setup work in the Assert section.
10. Check err first, then the result. If err is expected, assert its type/message/unwrap behavior.
11. Use table-driven tests for many variants of the same behavior. 
12. Avoid tables when each case needs a large unique setup.
13. Use t.Run subtests for table-driven tests.
14. In table cases, store only what differs. 
15. Put shared setup outside the table.
16. Keep setup minimal. Create only the inputs and dependencies required for the behavior under test.
17. Use stubs/fakes instead of real integrations. No network/DB/filesystem calls.
18. If something requires a real integration, ALWAYS ask user what to do and STOP.
19. Make tests deterministic: inject time (Clock/now), control randomness (fixed seed or constants), avoid flaky concurrency. 
20. Assert only what matters to the contract. Don’t snapshot huge structures unless the snapshot is the contract. 
21. Use helpers to reduce noise, but don’t hide the scenario. Mark helpers with t.Helper(). 
22. Don’t share mutable state across subtests. Each subtest gets fresh inputs and fakes. 
23. Use t.Parallel() only for fully isolated tests (no globals, no shared state, no order dependence). 
24. Make failures diagnostic: use `require` for error checks, setup steps, and conditions needed to proceed (fail fast). Use `assert` for validating business logic values and outcomes, and include context in assertions (expected vs actual).

## Execution
25. Run `make test` after changes. If it is too slow, ask the user before running a narrower test command.
