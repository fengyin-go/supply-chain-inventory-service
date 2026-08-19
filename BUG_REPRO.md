# Bug Reproduction

- Bug: Concurrent stocking can create duplicate inventory artifacts.
- Trigger: Submit two stocking requests for the same passed inspection at the same time.
- Symptoms: Both requests succeed and duplicate batches and movements are created.
