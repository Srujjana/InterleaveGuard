# InterleaveGuard: Concurrency Domain Knowledge

## Project context
This Go project contains concurrent programs using goroutines and shared state.
Analyze functions with attention to their behavior when executed concurrently.

## Shared-state assumptions
- Package-level variables may be accessed by multiple goroutines.
- Shared-state operations include READ, WRITE, and READ_WRITE.
- A READ_WRITE operation may represent a non-atomic read-modify-write sequence.
- Conflicting accesses to the same shared resource may cause data races or higher-level atomicity violations.

## Synchronization assumptions
- Mutex and RWMutex operations establish mutual-exclusion constraints.
- Shared-state accesses inside a critical section should be considered protected by the corresponding lock.
- Channel sends and receives may establish synchronization and ordering.
- WaitGroup operations coordinate goroutine completion.
- Do not assume that two operations are mutually exclusive merely because they occur in different functions.

## Specification-generation guidance
For each relevant function, infer:
- Preconditions on shared state and synchronization state.
- Postconditions describing the expected state change.
- Shared variables read or modified.
- Required lock ownership or synchronization assumptions.
- Whether a sequence of operations must behave atomically.
- Any assumptions about callers and concurrent callers.

Pay particular attention to:
- Unsynchronized shared-state access.
- Read-modify-write operations.
- Check-then-act sequences.
- Lock ordering and potential deadlocks.
- Operations whose correctness depends on another goroutine's actions.

## Important limitations
- Do not invent synchronization that is absent from the source code.
- Distinguish observed behavior from inferred intended behavior.
- Do not claim that a function is thread-safe solely because it uses a mutex somewhere.
- Treat inferred specifications as candidate contracts that require validation.
