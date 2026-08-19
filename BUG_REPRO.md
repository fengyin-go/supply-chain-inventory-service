# Bug Reproduction

- Bug: Concurrent outbound requests can pass the stock check together.
- Trigger: Submit two outbound requests for 8 units when only 10 units are available.
- Symptoms: Both requests succeed and the movement ledger records both requests.
