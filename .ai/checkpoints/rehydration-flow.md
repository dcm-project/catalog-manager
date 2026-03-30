# Rehydration Flow - Implementation Checkpoint

**Branch:** `rehydration-flow`
**Date:** 2026-03-31
**Status:** Implementation complete, all tests passing, updated to latest PM fork API

## Reference

- Enhancement spec: https://github.com/dcm-project/enhancements/blob/1f357c1213ccfbb8638f9b5baed82ada86114c15/enhancements/rehydration-flow/rehydration-flow.md
- Placement Manager fork: https://github.com/ygalblum/dcm-placement-manager/tree/rehydration-flow (commit `e9bd4542f0e1`)

## Commits

1. `d747be6` - **feat: add rehydrate endpoint for CatalogItemInstance**
   - Full implementation across all layers
2. `463d505` - **test: add unit tests for rehydrate endpoint**
   - Store, service, and handler unit tests
3. `fade03f` - **chore: update PM fork and rename new_instance_id to new_resource_id**
   - Updated `replace` directive to latest fork commit (`e9bd4542f0e1`)
   - Renamed `NewInstanceId` → `NewResourceId` in placement client

## What Was Implemented

### New API Endpoint

`POST /catalog-item-instances/{catalogItemInstanceId}:rehydrate`

- No request body
- Returns `200` with updated `CatalogItemInstance` (new `resource_id`)
- Returns `404` if instance not found
- Returns `500` on placement manager or internal failure

### Flow

1. Look up existing `CatalogItemInstance` by ID
2. Generate a new resource ID (UUID)
3. Call Placement Manager `POST /resources/{resourceId}:rehydrate` with `{"new_resource_id": "<newResourceID>"}`
4. Update `resource_id` in local DB
5. Return updated instance

### Files Modified

| File | Change |
|------|--------|
| `go.mod` / `go.sum` | `replace` directive pointing to PM fork (`ygalblum/dcm-placement-manager@rehydration-flow`, commit `e9bd4542f0e1`) |
| `api/v1alpha1/openapi.yaml` | New `:rehydrate` custom action path |
| `api/v1alpha1/spec.gen.go` | Regenerated |
| `internal/api/server/server.gen.go` | Regenerated - new `RehydrateCatalogItemInstance` on `StrictServerInterface` |
| `pkg/client/client.gen.go` | Regenerated - new client method |
| `internal/placement/client.go` | Added `RehydrateResource` to `Client` interface + implementation |
| `internal/service/errors.go` | Added `ErrPlacementManagerRehydrateFailed` |
| `internal/store/catalog_item_instance.go` | Added `UpdateResourceID` to `CatalogItemInstanceStore` interface + implementation |
| `internal/service/catalog_item_instance.go` | Added `Rehydrate` to `CatalogItemInstanceService` interface + implementation |
| `internal/handlers/v1alpha1/catalog_item_instance.go` | Added `RehydrateCatalogItemInstance` handler |
| `internal/handlers/v1alpha1/catalog_item_instance_errors.go` | Added `mapRehydrateCatalogItemInstanceErrorToHTTP` |
| `test/subsystem/setup_test.go` | WireMock stubs for PM rehydrate (success + failure) |
| `test/subsystem/catalog_item_instance_test.go` | 3 subsystem tests (success, 404, PM failure) |
| `internal/store/catalog_item_instance_test.go` | 2 unit tests for `UpdateResourceID` |
| `internal/service/catalog_item_instance_test.go` | 3 unit tests for `Rehydrate` |
| `internal/handlers/v1alpha1/catalog_item_instance_test.go` | 4 unit tests (1 success + 3 error mapping) |

### Test Results

- **Handlers Suite:** 65 specs - PASS
- **Placement Suite:** 6 specs - PASS
- **Service Suite:** 129 specs - PASS
- **Store Suite:** 48 specs - PASS
- **Subsystem tests:** compile OK (require Docker environment to run)

## Design Decisions

- **Dedicated `UpdateResourceID` store method** rather than extending existing `Update` (which only touches `display_name`, `spec`, `spec_catalog_item_id`) to avoid changing semantics for other callers.
- **No rollback on DB failure after PM success** - matches the existing pattern in `Delete`. Rehydrate is idempotent so can be retried.
- **operationId uses `:RehydrateCatalogItemInstance`** (colon prefix) per AEP custom action convention, as set by the linter.

## Open Items

- The `go.mod` `replace` directive points to a fork. This should be updated to point to the upstream `dcm-project/placement-manager` once the rehydration-flow changes are merged there.
