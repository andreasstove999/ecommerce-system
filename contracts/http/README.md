# HTTP Contracts

This folder contains the canonical **frontend contract** for the API Gateway (BFF).

- Contracts are versioned; see the `bff/v1` folder for the current OpenAPI definition.
- The gateway DTOs in `services/api-gateway-go/internal/http/dto` mirror this contract.
- OpenAPI is the source of truth; DTOs are generated/maintained to match it.

> Suggested (comment-only) generation for frontend typings:
>
> ```bash
> npx openapi-typescript ./contracts/http/bff/v1/openapi.yaml -o ./frontend/src/api.ts
> ```

## Visualization

You can visualize the OpenAPI contract using Swagger UI.

1. Start the UI service:
   ```bash
   cd ../../docker
   docker compose up -d swagger-ui
   ```
2. Open [http://localhost:8090](http://localhost:8090) in your browser.

