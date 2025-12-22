# Event Contracts

This folder contains example event envelopes and their JSON Schemas.

## Validating examples

Use the provided Make target to validate all JSON examples against their enveloped schemas:

```bash
make contracts-validate
```

The command installs pinned dependencies for the contracts package (using `npm`), walks every JSON file under `contracts/examples/**`, resolves the schema referenced by the example's `schema` field, and validates the full envelope with AJV. Validation fails with descriptive errors when an example does not satisfy its schema.

If you prefer not to use `make`, run the validator directly:

```bash
npm --prefix contracts install --no-fund --no-audit
npm --prefix contracts run validate
```
