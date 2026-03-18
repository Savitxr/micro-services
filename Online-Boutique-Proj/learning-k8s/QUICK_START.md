# QUICK_START.md: Get Started in 5 Minutes

## TL;DR

```bash
cd learning-k8s
docker-compose up --build
open http://localhost:8080
```

Then in another terminal:
```bash
# Test ProductCatalog
grpcurl -plaintext localhost:3550 hipstershop.ProductCatalogService/ListProducts

# Test CartService
grpcurl -plaintext -d '{"user_id":"test"}' localhost:7070 hipstershop.CartService/GetCart
```

---

##Step 1: Prerequisites (2 min)

Ensure you have:
- ✓ Docker Desktop installed (includes docker-compose)
- ✓ (Optional) grpcurl: `brew install grpcurl` or see [LOCAL_TESTING.md](docs/LOCAL_TESTING.md)

**Verify**:
```bash
docker --version
docker-compose --version
```

---

## Step 2: Build & Start (2 min)

```bash
cd learning-k8s

# Build and start all services
docker-compose up --build

# Or just start (if already built)
docker-compose up

# Output should show:
# - postgres is healthy
# - product-catalog listening on :3550
# - cart listening on :7070
# - frontend listening on :8080
```

**In a new terminal**, verify services:
```bash
docker-compose ps

# Expected: All services "Up"
```

---

## Step 3: Access Services (1 min)

### Frontend UI
```bash
open http://localhost:8080  # macOS
# or
curl http://localhost:8080  # See HTML response
```

### ProductCatalogService (gRPC)
```bash
# List all products
grpcurl -plaintext localhost:3550 hipstershop.ProductCatalogService/ListProducts

# Or if grpcurl not installed, test with docker
docker-compose exec product-catalog echo "Product Catalog running"
```

### CartService (gRPC)
```bash
# Get empty cart for user
grpcurl -plaintext -d '{"user_id":"customer-123"}' \
  localhost:7070 hipstershop.CartService/GetCart

# Response: empty cart for new user
```

### Database (PostgreSQL)
```bash
# Count products in database
docker exec learning-k8s-postgres psql -U postgres -d boutique_db \
  -c "SELECT COUNT(*) FROM products;"

# Expected: 9 products
```

---

## Step 4: Test Full Workflow (Optional)

### Add Item to Cart
```bash
grpcurl -plaintext -d '{
  "user_id": "my-customer",
  "item": {"product_id": "OLJCESPC7Z", "quantity": 1}
}' localhost:7070 hipstershop.CartService/AddItem
```

### Get User Cart
```bash
grpcurl -plaintext -d '{"user_id":"my-customer"}' \
  localhost:7070 hipstershop.CartService/GetCart

# Response shows 1 item (Sunglasses)
```

### Browse Frontend
1. Go to http://localhost:8080
2. See all products
3. Add items to cart via web UI
4. Cart persists across page refreshes
5. Stop services, then `docker-compose up` again
6. Cart data still there! (Persisted to PostgreSQL)

---

## Troubleshooting Quick Reference

| Problem | Solution |
|---------|----------|
| Port already in use | `docker-compose down; docker-compose up` |
| Services won't start | `docker-compose logs postgres` to see errors |
| "Can't connect" | Wait 5-10 sec for services to be ready |
| grcurl not found | Install with `brew install grpcurl` or skip (optional) |
| "No container" | Run `docker-compose up` first |

---

## What's Next?

- ✓ Services running locally
- ✓ Data persists to PostgreSQL
- ✓ gRPC communication works

**Next Steps**:
1. Read [README.md](README.md) for full project overview
2. Review [DEPLOYMENT.md](docs/DEPLOYMENT.md) to understand Docker images
3. Read [SERVICE_ARCHITECTURE.md](docs/SERVICE_ARCHITECTURE.md) for system design
4. Deploy to Kubernetes (Phase 2)

---

## Common Commands

```bash
# Cleanup (stop all, keep volumes)
docker-compose stop

# Full cleanup (remove everything)
docker-compose down -v

# View logs
docker-compose logs -f

# Enter container shell
docker-compose exec product-catalog sh

# Restart one service
docker-compose restart cart-service

# Rebuild one service
docker-compose build --no-cache product-catalog
```

---

## Documentation Map

| File | Purpose |
|------|---------|
| [README.md](README.md) | Project overview & getting started |
| [DEPLOYMENT.md](docs/DEPLOYMENT.md) | Docker builds & optimization |
| [NETWORKING.md](docs/NETWORKING.md) | gRPC service discovery |
| [ENVIRONMENT_VARIABLES.md](docs/ENVIRONMENT_VARIABLES.md) | All config reference |
| [VOLUMES_AND_STORAGE.md](docs/VOLUMES_AND_STORAGE.md) | Database persistence |
| [SERVICE_ARCHITECTURE.md](docs/SERVICE_ARCHITECTURE.md) | System design |
| [LOCAL_TESTING.md](docs/LOCAL_TESTING.md) | Complete testing guide |

---

**You're ready! Run `docker-compose up` and start exploring!**
