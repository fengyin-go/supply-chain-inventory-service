# Bug Reproduction

- Bug: Concurrent inbound creation is not atomic per purchase order.
- Trigger: Submit two inbound creation requests for the same confirmed purchase order at the same time.
- Symptoms: Both requests succeed and two inbound orders are stored.
