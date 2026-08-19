# Bug Reproduction

- Bug: Return stock deduction is not compensated when return status persistence fails.
- Trigger: Complete a pending return while the return writer is unavailable.
- Symptoms: Stock decreases although the return remains pending.
