# Contributing to Event Contracts

This pull request introduces the shared contracts only; service-level adoption will happen in a later PR. Use these guidelines to evolve contracts safely.

## Ownership

| Event | Owning team/service |
| ----- | ------------------- |
| EventEnvelope | Platform/Architecture |
| CartCheckedOut | Cart service |
| OrderCreated | Order service |
| OrderCompleted | Order service |
| PaymentSucceeded | Payment service |
| PaymentFailed | Payment service |
| StockReserved | Inventory service |
| StockDepleted | Inventory service |
| ShippingCreated | Shipping service |

Owners approve changes to their events and keep examples authoritative.

## Adding a new version

1. Start from the latest version of the event. Copy the payload and enveloped schemas, increment the version suffix (e.g., `OrderCreated.v1` → `OrderCreated.v2`), and keep the old versions intact.
2. Document the change in `contracts/CHANGELOG.md`, including rationale and compatibility expectations.
3. Update or add the complete example in `contracts/examples/<domain>/<EventName>.vX.json` to reflect the new contract.
4. Run `contracts-validate` and ensure it passes.

## Backward-compatibility policy

- Prefer additive changes (new optional fields) within a major version when possible.
- Breaking changes require a new version number and new schema files; never edit existing versions in place.
- Producers should continue emitting older versions until all consumers confirm readiness for the newer contract.
- Consumers must remain tolerant by accepting both the current and prior versions during transitions.

## Quality checks

- Schemas and examples must stay in sync for every versioned contract.
- CI must pass `contracts-validate` before merging.
- Include a brief summary of the contract change and consumer impact in the pull request description.
