# Bug review

The buggy use case breaks both the business rules and the architecture:

**Input validation**
1. No `nil` check on `req`, so it can panic.
2. Amount isn't checked for being positive.
3. Empty account IDs aren't validated.
4. Same source and destination isn't rejected.

**Error handling**
5. Both `Retrieve` errors are discarded.
6. A failed retrieval leaves `source` or `dest` nil, so a nil-pointer panic follows.

**Domain bypass**
7. Private state is changed directly instead of via `Withdraw` and `Deposit`.
8. That skips the insufficient-balance check, so the source can go negative.
9. It also skips status checks, so frozen accounts can send or receive.
10. `ChangeTracker` is bypassed, so dirty fields aren't recorded.

**Architecture**
11. The use case builds persistence mutations itself. That's the job of `AccountRepository.UpdateMut`.
12. Hand-built mutations always write `balance` instead of exactly the dirty fields.
13. It applies mutations itself, but it should only return a `Plan`. The service/committer boundary is skipped.
14. It returns only `error`, so it can't return the plan.

**Atomicity**
15. The two `Apply` calls aren't atomic. If the first succeeds and the second fails, only half the transfer is saved.
16. There's no rollback or recovery for that case.

The fix: validate the full request, retrieve both accounts, call the domain methods, get dirty-field mutations from the repository, and return both in one plan without applying them. The outer service then passes the plan to a transaction-aware committer that applies it atomically.