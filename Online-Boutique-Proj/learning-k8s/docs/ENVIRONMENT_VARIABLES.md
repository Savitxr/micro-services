# ENVIRONMENT_VARIABLES.md: Complete Reference

## Overview
All microservices and infrastructure components are configured via environment variables. This document provides a complete reference for development (docker-compose) and production (Kubernetes) environments.

---

## ProductCatalogService (Go)

### Database Connection

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `DB_HOST` | `localhost` | ✓ | PostgreSQL hostname |
| `DB_PORT` | `5432` | ✓ | PostgreSQL port |
| `DB_USER` | N/A | ✓ | Database user |
| `DB_PASSWORD` | N/A | ✓ | Database password |
| `DB_NAME` | `boutique_db` | ✓ | Database name |
| `DB_SSLMODE` | `disable` | - | SSL mode (disable/require/etc) |

### Service Configuration

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `PORT` | `3550` | - | gRPC listening port |
| `ENABLE_TRACING` | `0` | - | Enable OpenTelemetry tracing |
| `ENABLE_PROFILER` | `0` | - | Enable CPU/memory profiler |

### Docker Compose Example

```yaml
product-catalog:
  environment:
    PORT: 3550
    DB_HOST: postgres
    DB_PORT: 5432
    DB_USER: postgres
    DB_PASSWORD: postgres
    DB_NAME: boutique_db
```

### Kubernetes Example (Future)

```yaml
env:
- name: DB_HOST
  valueFrom:
    configMapKeyRef:
      name: product-catalog-config
      key: db-host
- name: DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: postgres-credentials
      key: password
```

---

## CartService (C# / .NET)

### Database Connection

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `ConnectionStrings__DefaultConnection` | N/A | ✓ | Full connection string |

**Format**:
```
Server=<host>;Port=<port>;Database=<db>;User Id=<user>;Password=<password>;
```

### Service Configuration

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `ASPNETCORE_URLS` | `http://+:5000` | - | Listening URLs |
| `ASPNETCORE_ENVIRONMENT` | `Production` | - | Environment (Development/Production) |
| `ASPNETCORE_Kestrel__Endpoints__Http__Url` | N/A | - | HTTP endpoint URL |

### Docker Compose Example

```yaml
cart-service:
  environment:
    ASPNETCORE_URLS: http://+:7070
    ConnectionStrings__DefaultConnection: "Server=postgres;Port=5432;Database=boutique_db;User Id=postgres;Password=postgres;"
```

### Kubernetes Example (Future)

```yaml
env:
- name: ASPNETCORE_URLS
  value: "http://+:7070"
- name: ConnectionStrings__DefaultConnection
  valueFrom:
    secretKeyRef:
      name: cart-service-secrets
      key: db-connection-string
```

---

## Frontend (Go)

### Service Discovery

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `PRODUCT_CATALOG_SERVICE_ADDR` | N/A | ✓ | ProductCatalogService gRPC address |
| `CART_SERVICE_ADDR` | N/A | ✓ | CartService gRPC address |
| `CURRENCY_SERVICE_ADDR` | N/A | - | CurrencyService address (if enabled) |
| `RECOMMENDATION_SERVICE_ADDR` | N/A | - | RecommendationService address |
| `CHECKOUT_SERVICE_ADDR` | N/A | - | CheckoutService address |
| `SHIPPING_SERVICE_ADDR` | N/A | - | ShippingService address |
| `AD_SERVICE_ADDR` | N/A | - | AdService address |

### Service Configuration

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `PORT` | `8080` | - | HTTP listening port |
| `LISTEN_ADDR` | `0.0.0.0` | - | Listening address |
| `BASE_URL` | N/A | - | Frontend base URL |
| `ENV_PLATFORM` | N/A | - | Platform identifier |

### Docker Compose Example

```yaml
frontend:
  environment:
    PORT: 8080
    PRODUCT_CATALOG_SERVICE_ADDR: product-catalog:3550
    CART_SERVICE_ADDR: cart-service:7070
    CURRENCY_SERVICE_ADDR: currency-service:7000
    RECOMMENDATION_SERVICE_ADDR: recommendation-service:8081
    CHECKOUT_SERVICE_ADDR: checkout-service:5050
    SHIPPING_SERVICE_ADDR: shipping-service:50051
    AD_SERVICE_ADDR: ad-service:9555
```

### Kubernetes Example (Future)

```yaml
env:
- name: PRODUCT_CATALOG_SERVICE_ADDR
  value: "product-catalog.default.svc.cluster.local:3550"
- name: CART_SERVICE_ADDR
  value: "cart-service.default.svc.cluster.local:7070"
```

---

## PostgreSQL

### Initialization

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `POSTGRES_USER` | N/A | ✓ | Root database user |
| `POSTGRES_PASSWORD` | N/A | ✓ | Root user password |
| `POSTGRES_DB` | `postgres` | - | Default database name |
| `POSTGRES_INITDB_ARGS` | N/A | - | Extra args for initdb |

### Docker Compose Example

```yaml
postgres:
  environment:
    POSTGRES_USER: postgres
    POSTGRES_PASSWORD: postgres
    POSTGRES_DB: boutique_db
```

---

## Complete Docker Compose Configuration

Full `.env` file style (for reference):

```bash
# PostgreSQL
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=boutique_db

# ProductCatalogService
PRODUCT_CATALOG_PORT=3550
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=boutique_db

# CartService
CART_PORT=7070
ASPNETCORE_URLS=http://+:7070

# Frontend
FRONTEND_PORT=8080
PRODUCT_CATALOG_SERVICE_ADDR=product-catalog:3550
CART_SERVICE_ADDR=cart-service:7070
CURRENCY_SERVICE_ADDR=currency-service:7000
RECOMMENDATION_SERVICE_ADDR=recommendation-service:8081
CHECKOUT_SERVICE_ADDR=checkout-service:5050
SHIPPING_SERVICE_ADDR=shipping-service:50051
AD_SERVICE_ADDR=ad-service:9555
```

---

## Environment-Specific Values

### Docker Compose (Local Development)

```
Service Discovery: Use container names (service-to-service)
  PRODUCT_CATALOG_SERVICE_ADDR=product-catalog:3550
  
Database Connection: container name as hostname
  DB_HOST=postgres
  
Ports: 127.0.0.1:PORT accessible to host
  Frontend: http://localhost:8080
```

### Kubernetes (Production)

```
Service Discovery: Use Kubernetes DNS (fully qualified)
  PRODUCT_CATALOG_SERVICE_ADDR=product-catalog.default.svc.cluster.local:3550
  
Database Connection: Kubernetes service DNS
  DB_HOST=postgres-service.default.svc.cluster.local
  
Ports: Ingress or NodePort routes external traffic
  Frontend: https://boutique.example.com
```

---

## ConfigMaps vs Secrets (Kubernetes Future State)

### ConfigMap (Non-Sensitive Data)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: product-catalog-config
data:
  db-host: postgres-service
  db-port: "5432"
  port: "3550"
```

### Secret (Sensitive Data)

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: postgres-credentials
type: Opaque
data:
  username: <base64-encoded>
  password: <base64-encoded>
```

---

## Troubleshooting

| Issue | Check |
|-------|-------|
| Service not connecting to database | `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD` correct |
| Frontend shows "service unavailable" | `PRODUCT_CATALOG_SERVICE_ADDR`, `CART_SERVICE_ADDR` correct |
| Services not starting | Check for typos in env var names (case-sensitive in Linux) |
| Connection string errors | `.NET`: use `__` for nesting; PostgreSQL: no spaces in connection string |

---

## Migration: Docker Compose → Kubernetes

When moving from docker-compose to K8s:

1. **Service Names Change**:
   - Old: `product-catalog:3550`
   - New: `product-catalog.default.svc.cluster.local:3550`

2. **Database Host Changes**:
   - Old: `postgres`
   - New: `postgres-service.default.svc.cluster.local` or external RDS endpoint

3. **Secrets Handling**:
   - Old: Plain text in docker-compose.yml
   - New: Kubernetes Secrets + RBAC

4. **ConfigMaps**:
   - Old: environment: inline in docker-compose.yml
   - New: Kubernetes ConfigMaps + env refs

See `VOLUMES_AND_STORAGE.md` for persistence strategy changes.
