# Bug Reproduction

- Bug: Inspection status and inspection record are committed separately.
- Trigger: Start inspection while the inspection writer returns an error.
- Symptoms: The inbound order becomes inspecting without a matching inspection record.
