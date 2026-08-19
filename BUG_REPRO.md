# Bug Reproduction

- Bug: Purchase and inbound items share mutable state.
- Trigger: Create an inbound order, mutate the loaded purchase item quantity, then complete inspection and stocking.
- Symptoms: Stock is inflated and an over-limit return is accepted.
