# Bug Reproduction

- Bug: Pending returns do not reserve inbound quantity.
- Trigger: Create two pending returns of 6 units against an inbound quantity of 10.
- Symptoms: Both return orders are accepted and the second later cannot complete.
