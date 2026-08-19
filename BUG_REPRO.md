# Bug Reproduction

- Bug: Exported snapshots expose internal entity pointers and slices.
- Trigger: Mutate exported product and purchase item values, then read the service state again.
- Symptoms: Stored product data and report values change.
