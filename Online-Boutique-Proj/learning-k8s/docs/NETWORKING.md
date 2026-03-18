# NETWORKING.md: gRPC Service Discovery & Communication

## Overview
Services communicate via gRPC (HTTP/2 protocol) with strong typing through Protocol Buffers. This document explains service discovery, port configuration, and network communication patterns.

---

## Service Port Configuration

| Service | Port | Protocol | Purpose |
|---------|------|----------|---------|
| ProductCatalogService | 3550 | gRPC | Internal backend communication |
| CartService | 7070 | gRPC | Internal backend communication |
| Frontend | 8080 | HTTP/REST | User-facing web interface |
| PostgreSQL | 5432 | TCP | Database connections |

---

## Service Discovery

### Docker Compose (Local Development)

Services discover each other by **container name** as DNS hostname:

```
product-catalog    → address: product-catalog:3550
cart-service       → address: cart-service:7070
postgres           → address: postgres:5432
frontend           → address: frontend:8080
```

**How It Works**:
- Docker creates internal DNS resolver for named containers
- Service names resolve to container IP within the network `learning-k8s-net`
- No manual IP management needed

**Example Frontend Code**:
```python
# Frontend (main.go)
productCatalogSvcAddr := os.Getenv("PRODUCT_CATALOG_SERVICE_ADDR")  # "product-catalog:3550"
conn, _ := grpc.NewClient(productCatalogSvcAddr, ...)
```

### Kubernetes Service Discovery (Future Phase)

In Kubernetes, services are discovered via **Kubernetes DNS**:

```
product-catalog.default.svc.cluster.local:3550
cart-service.default.svc.cluster.local:7070
```

Format: `<service-name>.<namespace>.svc.cluster.local:<port>`

---

## gRPC Service Contracts

All backend services use Protocol Buffers defined in `protos/demo.proto`.

### ProductCatalogService (gRPC)

**Port**: 3550

**Service Definition** (from demo.proto):
```proto
service ProductCatalogService {
    rpc ListProducts(Empty) returns (ListProductsResponse) {}
    rpc GetProduct(GetProductRequest) returns (Product) {}
    rpc SearchProducts(SearchProductsRequest) returns (SearchProductsResponse) {}
}
```

**Service Address Environment Variable**:
```
PRODUCT_CATALOG_SERVICE_ADDR=product-catalog:3550
```

### CartService (gRPC)

**Port**: 7070

**Service Definition**:
```proto
service CartService {
    rpc AddItem(AddItemRequest) returns (Empty) {}
    rpc GetCart(GetCartRequest) returns (Cart) {}
    rpc EmptyCart(EmptyCartRequest) returns (Empty) {}
}
```

**Service Address Environment Variable**:
```
CART_SERVICE_ADDR=cart-service:7070
```

---

## Health Checks

### gRPC Health Check Protocol

All backend services implement `grpc.health.v1.Health` standard health check:

```proto
service Health {
  rpc Check(HealthCheckRequest) returns (HealthCheckResponse) {}
  rpc Watch(HealthCheckRequest) returns (stream HealthCheckResponse) {}
}
```

### Docker Compose Health Check

```yaml
healthcheck:
  test: ["CMD-SHELL", "pg_isready -U postgres"]
  interval: 5s
  timeout: 5s
  retries: 5
```

### Kubernetes Health Check (Future)

```yaml
livenessProbe:
  grpc:
    port: 3550
  initialDelaySeconds: 5
  periodSeconds: 10

readinessProbe:
  grpc:
    port: 3550
  initialDelaySeconds: 2
  periodSeconds: 5
```

---

## Network Topology (Docker Compose)

```
┌─────────────────────────────────────────────────────────┐
│ Docker Network: learning-k8s-net (bridge)              │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌──────────────┐                                      │
│  │  Frontend    │                                      │
│  │  :8080       │◄──── HTTP http://localhost:8080      │
│  └──────┬───────┘                                      │
│         │ gRPC (product-catalog:3550)                  │
│         │ gRPC (cart-service:7070)                     │
│         ├──────────────┬──────────────┐                │
│         │              │              │                │
│  ┌──────▼──────┐ ┌─────▼──────┐ ┌────▼─────────┐     │
│  │ ProductCat  │ │CartService │ │  PostgreSQL  │     │
│  │  :3550      │ │   :7070    │ │   :5432      │     │
│  └─────┬───────┘ └─────┬──────┘ └────┬─────────┘     │
│        │                │             │               │
│        └────────────┬───┘             │               │
│                     │ SQL TCP         │               │
│                     └─────────────────┘               │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## Network Communication Patterns

### gRPC Unary RPC (Request-Response)

1. Frontend initiates gRPC call to ProductCatalogService
2. Service handles request, queries PostgreSQL
3. Service returns response
4. Frontend receives response and processes

**Example**:
```
Frontend → ProductCatalogService.GetProduct(id: "123")
         → ProductCatalogService queries PostgreSQL
         → ProductCatalogService returns Product proto
Frontend → Receives product details, renders HTML
```

### Database Connections

1. ProductCatalogService: Maintains connection pool to PostgreSQL
2. CartService: Maintains connection pool to PostgreSQL
3. Connection strings set via environment variables

---

## Dependency Order

**Startup Sequence** (docker-compose handles this):

1. **PostgreSQL** boots first
   - Listen on :5432
   - Execute init.sql → schema creation
   - Execute seed-data.sql → product catalog population
   - Healthy when `pg_isready` returns true

2. **ProductCatalogService** boots (depends on PostgreSQL healthy)
   - Connect to postgres:5432
   - Listen on :3550 for gRPC
   - Ready for requests

3. **CartService** boots (depends on PostgreSQL healthy)
   - Connect to postgres:5432
   - Listen on :7070 for gRPC
   - Ready for requests

4. **Frontend** boots (depends on ProductCatalog + CartService running)
   - Connect to product-catalog:3550
   - Connect to cart-service:7070
   - Listen on :8080 for HTTP
   - Serve web interface

---

## Environment Variables for Discovery

See `ENVIRONMENT_VARIABLES.md` for complete reference. Key discovery variables:

```bash
# Docker Compose (Local)
PRODUCT_CATALOG_SERVICE_ADDR=product-catalog:3550
CART_SERVICE_ADDR=cart-service:7070

# Kubernetes (Future)
PRODUCT_CATALOG_SERVICE_ADDR=product-catalog.default.svc.cluster.local:3550
CART_SERVICE_ADDR=cart-service.default.svc.cluster.local:7070
```

---

## Troubleshooting Connection Issues

| Problem | Cause | Solution |
|---------|-------|----------|
| "Connection refused" | Service not running | Check `docker-compose ps` |
| "Name resolution failed" | Wrong hostname | Verify container name in docker-compose.yml |
| "Deadline exceeded" | Service too slow | Check logs: `docker-compose logs product-catalog` |
| "Connection reset" | Service crashed | Check service health: `docker-compose logs` |

---

## Security Notes

- **Current Setup**: Services communicate within Docker bridge network (isolated from host)
- **No TLS/mTLS**: Development configuration only
- **Future**: Implement mTLS for production K8s deployments
- **Database**: Unencrypted connection (development only); use SSL in production

---

## Next Steps

After understanding local networking:
1. Learn Kubernetes Service abstraction
2. Implement Kubernetes NetworkPolicies for more restrictive routing
3. Add mTLS/TLS between services
4. Implement circuit breakers and service mesh (Istio/Linkerd)
