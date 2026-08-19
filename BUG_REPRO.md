# Bug Reproduction

- Bug: Multi-item inbound stocking leaves earlier artifacts after a later movement write fails.
- Trigger: Stock an inbound order with two items while the second movement write fails.
- Symptoms: The first batch remains and the inbound order stays inspecting.
