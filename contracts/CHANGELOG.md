# Event Contract Changelog

This repository currently introduces the shared event contracts only; individual services will adopt these contracts in a follow-up pull request.

## Baseline versions

All events start at **v1** in this PR. Future changes must append new versions rather than mutate existing schemas.

| Domain | Event | Version | Notes |
| ------ | ----- | ------- | ----- |
| envelope | EventEnvelope | v1 | Initial shared envelope for all contracts. |
| cart | CartCheckedOut | v1 | Initial contract for cart checkout events. |
| order | OrderCreated | v1 | Initial contract emitted when an order is created. |
| order | OrderCompleted | v1 | Initial contract emitted when an order is completed. |
| payment | PaymentSucceeded | v1 | Initial contract for successful payment captures. |
| payment | PaymentFailed | v1 | Initial contract for failed payment captures. |
| inventory | StockReserved | v1 | Initial contract for reserving inventory against an order. |
| inventory | StockDepleted | v1 | Initial contract for communicating inventory depletion. |
| shipping | ShippingCreated | v1 | Initial contract for new shipment creation. |

## How to record future changes

When creating a new version of an event:
- Copy the latest payload and enveloped schemas to new files with the incremented version (e.g., `OrderCreated.v2.*.json`).
- Document the change in this changelog with rationale and compatibility notes.
- Update the corresponding example under `contracts/examples/<domain>/<EventName>.vX.json`.
