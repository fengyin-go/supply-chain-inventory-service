# Bug Reproduction

- Bug: Outbound batch deduction is not restored after movement persistence fails.
- Trigger: Submit an outbound request while the outbound movement writer is unavailable.
- Symptoms: The request fails but stock remains reduced.
