# Learning K8s: 3-Tier Microservices Project

A simplified, production-ready microservices application for learning Kubernetes fundamentals. This project extracts the essentials from Google Cloud's Online Boutique to create a focused learning environment.

---

## Quick Start

```bash
cd learning-k8s

# Build and start all services
docker-compose up --build

# In another terminal, verify services are running
docker-compose ps

# Access Frontend
open http://localhost:8080

# Test ProductCatalogService
grpcurl -plaintext localhost:3550 hipstershop.ProductCatalogService/ListProducts

# Test CartService
grpcurl -plaintext -d '{"user_id":"test-user"}' localhost:7070 hipstershop.CartService/GetCart
```

---

## Project Structure

```
learning-k8s/
├── services/
│   ├── product-catalog/       Go service + PostgreSQL
│   └── cart/                  C# / .NET service + PostgreSQL
├── frontend/                  Go HTTP service (unmodified)
├── database/
│   ├── init.sql               Schema creation
│   └── seed-data.sql          Product catalog seed
├── protos/                    gRPC service contracts
├── docs/
│   ├── DEPLOYMENT.md          Docker image building
│   ├── NETWORKING.md          gRPC service discovery
│   ├── ENVIRONMENT_VARIABLES.md   Complete env reference
│   ├── VOLUMES_AND_STORAGE.md     Database persistence & K8s PVCs
│   ├── SERVICE_ARCHITECTURE.md    System design & patterns
│   └── LOCAL_TESTING.md       Verification procedures
├── docker-compose.yml         Local development orchestration
└── README.md                  This file
```

---

## Architecture Highlights

### Services

| Service | Language | Purpose | State |
|---------|----------|---------|-------|
| **ProductCatalogService** | Go | Product queries, search | Stateless (queries PostgreSQL) |
| **CartService** | C# / .NET | User shopping carts | Stateless (persists to PostgreSQL) |
| **Frontend** | Go | User-facing HTTP/web UI | Stateless |
| **PostgreSQL** | SQL | Persistent data storage | Stateful (volumes) |

### Key Design Principles

- **No inter-service dependencies**: ProductCatalog ↔ CartService are independent
- **Stateless services**: All state in PostgreSQL; services easily replaceable
- **gRPC communication**: Efficient, strongly-typed internal APIs
- **Multi-language**: Go, C# / .NET demonstrate polyglot containerization
- **Database-backed**: Both services query PostgreSQL; easy to scale

---

## System Diagram

```
User Browser (HTTP :8080)
    ↓
Frontend (Go)
    ├─→ ProductCatalogService (gRPC :3550)
    │        ↓
    │   PostgreSQL (TCP :5432)
    │
    └─→ CartService (C# :7070)
         ↓
    PostgreSQL (TCP :5432)
```

**No direct ProductCatalog → CartService calls** (no cascading failures)

---

## Getting Started

### Prerequisites

- Docker Desktop (with docker-compose)
- grpcurl (optional, for gRPC testing): `brew install grpcurl`
- curl (for HTTP testing)

### 1. Build Services

```bash
cd learning-k8s
docker-compose build
```

**Output**: 4 Docker images built
- `learning-k8s_postgres`
- `learning-k8s_product-catalog`
- `learning-k8s_cart-service`
- `learning-k8s_frontend`

### 2. Start Services

```bash
docker-compose up -d
```

**Services Start Order**:
1. PostgreSQL starts, initializes schema, loads seed data
2. ProductCatalogService connects to DB, listens on :3550
3. CartService connects to DB, listens on :7070
4. Frontend connects to both services, listens on :8080

### 3. Verify Services

```bash
docker-compose ps

# Expected: All services "Up"
```

### 4. Test Locally

See [LOCAL_TESTING.md](docs/LOCAL_TESTING.md) for complete testing procedures.

**Quick Test**:
```bash
# Frontend UI
curl http://localhost:8080

# ProductCatalog gRPC
grpcurl -plaintext localhost:3550 hipstershop.ProductCatalogService/ListProducts

# CartService gRPC  
grpcurl -plaintext -d '{"user_id":"test"}' localhost:7070 hipstershop.CartService/GetCart
```

---

## Documentation Guide

| Document | Purpose | Read When |
|----------|---------|-----------|
| [DEPLOYMENT.md](docs/DEPLOYMENT.md) | Docker image building & optimization | Customizing containers, understanding multi-stage builds |
| [NETWORKING.md](docs/NETWORKING.md) | gRPC service discovery & communication | Configuring service mesh, adding new services |
| [ENVIRONMENT_VARIABLES.md](docs/ENVIRONMENT_VARIABLES.md) | Complete environment reference | Deploying to K8s, exporting configs |
| [VOLUMES_AND_STORAGE.md](docs/VOLUMES_AND_STORAGE.md) | Database persistence & K8s PVCs | Understanding stateful workloads, backup/disaster recovery |
| [SERVICE_ARCHITECTURE.md](docs/SERVICE_ARCHITECTURE.md) | System design, failure scenarios, scalability | Understanding architecture, designing K8s manifests |
| [LOCAL_TESTING.md](docs/LOCAL_TESTING.md) | Testing & verification procedures | Running & validating locally before K8s deployment |

---

## Key Learning Objectives

By completing this project, you will understand:

- ✓ **Microservices Architecture**: Independent services, no cascading failures
- ✓ **gRPC Communication**: Strongly-typed service contracts via Protocol Buffers
- ✓ **Containerization**: Docker multi-stage builds, image optimization
- ✓ **Service Discovery**: DNS-based routing (docker-compose → Kubernetes)
- ✓ **Persistent Storage**: Volumes, StatefulSets, backing databases
- ✓ **Health Checks**: liveness/readiness probes, graceful startup
- ✓ **Environment Configuration**: Env vars, ConfigMaps, Secrets
- ✓ **Failure Isolation**: One service crashes; others continue

---

## Common Workflows

### Add Item to Cart

```bash
grpcurl -plaintext -d '{
  "user_id": "customer-123",
  "item": {
    "product_id": "OLJCESPC7Z",
    "quantity": 2
  }
}' localhost:7070 hipstershop.CartService/AddItem
```

### Get User Cart

```bash
grpcurl -plaintext -d '{"user_id":"customer-123"}' \
  localhost:7070 hipstershop.CartService/GetCart
```

### Search Products

```bash
grpcurl -plaintext -d '{"query":"watch"}' \
  localhost:3550 hipstershop.ProductCatalogService/SearchProducts
```

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f product-catalog
docker-compose logs -f cart-service
docker-compose logs -f postgres
```

---

## What's NOT Included (By Design)

To keep focus on K8s fundamentals, these advanced features are intentionally excluded:

- ❌ Kubernetes manifests (next phase: create YAML)
- ❌ Other microservices (CheckoutService, PaymentService, etc.)
- ❌ Load generator
- ❌ OpenTelemetry/tracing (added on demand)
- ❌ Service mesh (Istio/Linkerd)
- ❌ API gateway

These can be added incrementally as you grow the project.

---

## Troubleshooting

### Services Won't Start

```bash
# Check logs
docker-compose logs postgres

# Common issues:
# - Port already in use: docker ps | grep :5432
# - Volume issues: docker volume ls
# - Image build failed: docker-compose build --no-cache
```

### Can't Connect to Services

```bash
# Verify running
docker-compose ps

# Test network connectivity
docker-compose exec frontend ping product-catalog

# Check service listening
docker-compose exec product-catalog ss -tlnp | grep 3550
```

### Data Lost After Restart

```bash
# You likely ran: docker-compose down -v
# This DELETES volumes (data)

# Use instead:
docker-compose down  # Keeps volumes
# or
docker-compose stop  # Keeps containers & volumes

# To preserve at startup:
docker-compose up   # Mounts existing volumes
```

---

## Next Steps (Kubernetes)

Once comfortable with this local setup:

1. **Create Kubernetes manifests** for each service
2. **Deploy to local K8s** (Minikube, Docker Desktop K8s, or Kind)
3. **Add ConfigMaps & Secrets** for configuration
4. **Implement StatefulSet** for PostgreSQL
5. **Set up Ingress** for Frontend access
6. **Add monitoring** (Prometheus, Grafana)

Templates and examples will be provided as part of Phase 2.

---

## Performance Targets

Expect on localhost (docker-compose):

| Operation | Target | Notes |
|-----------|--------|-------|
| ListProducts | <100ms | Depends on product count |
| GetProduct | <50ms | Single ID lookup |
| AddItem to Cart | <50ms | Database insert |
| GetCart | <30ms | User-specific query |
| Frontend page load | <500ms | HTML rendering + gRPC calls |

Actual times depend on machine specs. Monitor with:
```bash
docker stats
```

---

## Contributing & Customizing

### Add a New Product

```sql
INSERT INTO products (id, name, description, picture_url, price_usd_currency_code, price_usd_units, price_usd_nanos, categories)
VALUES ('MY_PRODUCT_ID', 'Product Name', 'Description', '/img/url.jpg', 'USD', 29, 990000000, ARRAY['category1', 'category2']);
```

### Modify CartService Connection String

Edit `docker-compose.yml`:
```yaml
environment:
  ConnectionStrings__DefaultConnection: "Server=postgres;..."
```

### Add Environment Variables

1. Update `docker-compose.yml` under `environment:`
2. Update corresponding service code to read the variable
3. Document in [ENVIRONMENT_VARIABLES.md](docs/ENVIRONMENT_VARIABLES.md)

---

## Resources

- [gRPC Web](https://grpc.io/blog/grpc-web-dance-with-protobuf/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [Docker Best Practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Online Boutique (Original Project)](https://github.com/GoogleCloudPlatform/microservices-demo)

---

## License

This learning project is based on Google Cloud's Online Boutique, an open-source project for learning cloud-native development.

---

## Support

For issues or questions:
1. Check the relevant documentation file in `docs/`
2. Review [LOCAL_TESTING.md](docs/LOCAL_TESTING.md) for common issues
3. Examine service logs: `docker-compose logs <service-name>`

---

**Ready to learn K8s? Start with `docker-compose up` and explore!**
