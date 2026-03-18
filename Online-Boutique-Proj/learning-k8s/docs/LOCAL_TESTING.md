# LOCAL_TESTING.md: Verification & Testing Procedures

## Overview
Complete step-by-step guide to build, run, and test all services locally using docker-compose.

---

## Prerequisites

- Docker Desktop (with docker-compose)
- grpcurl (for gRPC testing)
- curl (for HTTP testing)

### Install grpcurl (Optional but recommended)

```bash
# macOS
brew install grpcurl

# Windows (using Scoop)
scoop install grpcurl

# Linux
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# From source
git clone https://github.com/fullstorydev/grpcurl.git
cd grpcurl
go install ./cmd/grpcurl
```

---

## Phase 1: Build Images

### Clean Previous Build (Optional)

```bash
cd learning-k8s
docker-compose down -v  # Remove all containers and volumes
```

### Build All Images

```bash
cd learning-k8s
docker-compose build

# Output should show:
# Building postgres
# Building product-catalog
# Building cart-service
# Building frontend
```

**Verification**: Each service shows `Built` or `Already built`

### Build Individual Services (If Needed)

```bash
# Only ProductCatalogService
docker-compose build product-catalog

# Only CartService
docker-compose build cart-service

# Only Frontend
docker-compose build frontend
```

---

## Phase 2: Start Services

### Start All Services

```bash
cd learning-k8s
docker-compose up

# Output should show:
# learning-k8s-postgres     | ready to accept connections
# learning-k8s-product-catalog | listening on :3550
# learning-k8s-cart         | started
# learning-k8s-frontend     | listening on :8080
```

### Start in Background

```bash
docker-compose up -d

# Verify running services
docker-compose ps

# Expected output:
# NAME                      STATUS          PORTS
# learning-k8s-postgres     Up (healthy)    5432/tcp
# learning-k8s-product-catalog   Up         3550/tcp
# learning-k8s-cart         Up              7070/tcp
# learning-k8s-frontend     Up              8080/tcp
```

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f product-catalog
docker-compose logs -f cart-service
docker-compose logs -f postgres

# Last N lines
docker-compose logs --tail=50 postgres
```

---

## Phase 3: Verify Database

### Check PostgreSQL Connection

```bash
# Using psql inside container
docker exec learning-k8s-postgres psql -U postgres -d boutique_db -c "SELECT COUNT(*) FROM products;"

# Expected output:
# count
# -----
#     9
```

### Verify Schema

```bash
# List all tables
docker exec learning-k8s-postgres psql -U postgres -d boutique_db -c "\dt"

# Expected output:
#             List of relations
#  Schema |    Name     | Type  |  Owner
# --------+-------------+-------+----------
#  public | cart_items  | table | postgres
#  public | products    | table | postgres
#  public | user_carts  | table | postgres
```

### View Product Data

```bash
docker exec learning-k8s-postgres psql -U postgres -d boutique_db -c "SELECT id, name, price_usd_units FROM products LIMIT 3;"

# Expected output:
#       id       |         name         | price_usd_units
# ---------------+----------------------+-----------------
#  OLJCESPC7Z    | Sunglasses           |              19
#  66VCHSJNUP    | Tank Top             |              18
#  1YMWWN1N4O    | Watch                |             109
```

---

## Phase 4: Test ProductCatalogService (gRPC)

### Test: ListProducts

```bash
grpcurl -plaintext localhost:3550 hipstershop.ProductCatalogService/ListProducts

# Expected output:
# {
#   "products": [
#     {
#       "id": "OLJCESPC7Z",
#       "name": "Sunglasses",
#       "description": "Add a modern touch to your outfits with these sleek aviator sunglasses.",
#       ...
```

### Test: GetProduct

```bash
grpcurl -plaintext -d '{"id":"OLJCESPC7Z"}' localhost:3550 hipstershop.ProductCatalogService/GetProduct

# Expected output:
# {
#   "id": "OLJCESPC7Z",
#   "name": "Sunglasses",
#   "priceUsd": {
#     "currencyCode": "USD",
#     "units": 19,
#     "nanos": 990000000
#   },
#   ...
```

### Test: SearchProducts

```bash
grpcurl -plaintext -d '{"query":"watch"}' localhost:3550 hipstershop.ProductCatalogService/SearchProducts

# Expected output:
# {
#   "results": [
#     {
#       "id": "1YMWWN1N4O",
#       "name": "Watch",
#       ...
```

### Verify Service Health (gRPC)

```bash
grpcurl -plaintext localhost:3550 grpc.health.v1.Health/Check

# Expected output:
# {
#   "status": "SERVING"
# }
```

---

## Phase 5: Test CartService (gRPC)

### Test: AddItem to Cart

```bash
grpcurl -plaintext -d '{
  "user_id": "test-user-1",
  "item": {
    "product_id": "OLJCESPC7Z",
    "quantity": 2
  }
}' localhost:7070 hipstershop.CartService/AddItem

# Expected output: {} (empty response = success)
```

### Test: GetCart

```bash
grpcurl -plaintext -d '{"user_id":"test-user-1"}' localhost:7070 hipstershop.CartService/GetCart

# Expected output:
# {
#   "user_id": "test-user-1",
#   "items": [
#     {
#       "product_id": "OLJCESPC7Z",
#       "quantity": 2
#     }
#   ]
# }
```

### Test: Add Another Item

```bash
grpcurl -plaintext -d '{
  "user_id": "test-user-1",
  "item": {
    "product_id": "66VCHSJNUP",
    "quantity": 1
  }
}' localhost:7070 hipstershop.CartService/AddItem
```

### Test: Get Updated Cart

```bash
grpcurl -plaintext -d '{"user_id":"test-user-1"}' localhost:7070 hipstershop.CartService/GetCart

# Expected: 2 items (Sunglasses x2, Tank Top x1)
```

### Test: EmptyCart

```bash
grpcurl -plaintext -d '{"user_id":"test-user-1"}' localhost:7070 hipstershop.CartService/EmptyCart

# Expected output: {} (success)
```

### Test: Verify Cart is Empty

```bash
grpcurl -plaintext -d '{"user_id":"test-user-1"}' localhost:7070 hipstershop.CartService/GetCart

# Expected:
# {
#   "user_id": "test-user-1",
#   "items": []
# }
```

---

## Phase 6: Test Frontend (HTTP)

### Access Frontend UI

```bash
open http://localhost:8080  # macOS
# or
curl http://localhost:8080
```

**Expected**:
- Homepage loads with product catalog
- Products visible (Sunglasses, Tank Top, Watch, etc.)
- Add to cart buttons functional
- No error messages in browser console

### Frontend Testing Checklist

- [ ] Homepage loads (no connection errors)
- [ ] Products are visible with names, descriptions, prices
- [ ] Add to cart button works for each product
- [ ] Click "View Cart" shows items
- [ ] Cart persists after page refresh
- [ ] Remove item from cart works
- [ ] Empty cart button clears all items

### Browser Console Check

Open Developer Tools (F12) and check Console tab:
- No red error messages
- No "Service unavailable" warnings
- Network requests to localhost succeed

### Test Workflow

1. **Navigate** to http://localhost:8080
2. **Browse** product catalog
3. **Add** Sunglasses (quantity: 1)
4. **Add** Tank Top (quantity: 2)
5. **View** cart → should show 2 items
6. **Modify** Tank Top quantity to 3
7. **Remove** Tank Top
8. **Add** Watch
9. **Refresh** page → cart should persist
10. **Empty** cart → all items removed

---

## Phase 7: Persistence Testing

### Test 1: Data Persists After Container Restart

```bash
# 1. Add item to cart
grpcurl -plaintext -d '{
  "user_id": "persist-test",
  "item": {"product_id": "OLJCESPC7Z", "quantity": 1}
}' localhost:7070 hipstershop.CartService/AddItem

# 2. Restart CartService
docker-compose restart cart-service

# Wait ~3 seconds for restart

# 3. Verify data persists
grpcurl -plaintext -d '{"user_id":"persist-test"}' localhost:7070 hipstershop.CartService/GetCart

# Expected: Item still in cart (data persisted)
```

### Test 2: Data Survives Full Stack Restart

```bash
# 1. Add multiple items
grpcurl -plaintext -d '{"user_id":"full-test","item":{"product_id":"OLJCESPC7Z","quantity":1}}' localhost:7070 hipstershop.CartService/AddItem
grpcurl -plaintext -d '{"user_id":"full-test","item":{"product_id":"66VCHSJNUP","quantity":2}}' localhost:7070 hipstershop.CartService/AddItem

# 2. Stop all services
docker-compose down

# Wait ~2 seconds

# 3. Start again
docker-compose up -d

# Wait ~10 seconds for services to become healthy

# 4. Verify data persists
grpcurl -plaintext -d '{"user_id":"full-test"}' localhost:7070 hipstershop.CartService/GetCart

# Expected: Both items still exist
```

---

## Phase 8: Failure Simulation

### Test: ProductCatalog Service Restart

```bash
# Terminal 1: View logs
docker-compose logs -f product-catalog

# Terminal 2: Make request while service is running
grpcurl -plaintext localhost:3550 hipstershop.ProductCatalogService/ListProducts

# Terminal 3: Kill the container
docker-compose restart product-catalog

# In Terminal 2: grpcurl will show error (connection refused)
# Wait ~5 seconds, try again
grpcurl -plaintext localhost:3550 hipstershop.ProductCatalogService/ListProducts

# Expected: Service responds normally (auto-restarted)
```

### Test: Database Unavailability

```bash
# Stop PostgreSQL
docker-compose stop postgres

# Try ProductCatalog (will fail - no database)
grpcurl -plaintext localhost:3550 hipstershop.ProductCatalogService/ListProducts
# Expected: Connection timeout or error

# Restart PostgreSQL
docker-compose start postgres

# Wait ~5 seconds for health check
grpcurl -plaintext localhost:3550 hipstershop.ProductCatalogService/ListProducts

# Expected: Service responds normally
```

---

## Common Issues & Troubleshooting

| Issue | Cause | Solution |
|-------|-------|----------|
| "port 3550 already in use" | Another service on port | `lsof -i :3550` to find process; kill it or use different port |
| "database connection refused" | PostgreSQL not healthy | Check `docker-compose logs postgres` |
| "grpcurl: command not found" | grpcurl not installed | Install grpcurl or use `docker exec` with psql |
| "frontend shows blank" | ProductCatalog/Cart unavailable | Check `docker-compose ps` for service status |
| "image not found" | Build failed silently | Run `docker-compose build --no-cache` |

---

## Performance Testing

### LoadTesting ProductCatalog (Optional)

```bash
# Simple sequential requests
for i in {1..100}; do
  grpcurl -plaintext localhost:3550 hipstershop.ProductCatalogService/ListProducts > /dev/null &
done
wait

# Check response times in logs
docker-compose logs product-catalog | grep "duration"
```

### Monitor Resource Usage

```bash
# While running load test
docker stats learning-k8s-product-catalog learning-k8s-cart learning-k8s-postgres

# Typical metrics:
# - ProductCatalog: ~50MB RAM, <5% CPU
# - CartService: ~300MB RAM, <5% CPU  
# - PostgreSQL: ~200MB RAM, <10% CPU
```

---

## Cleanup

### Stop Services (Keep Volumes)

```bash
docker-compose stop
# Services stopped but data persists
# Start again with docker-compose start
```

### Full Cleanup (Remove Everything)

```bash
docker-compose down -v
# Removes: containers, networks, volumes (DATA DELETED)
# Fresh start on next docker-compose up
```

### Clean Build

```bash
docker-compose down -v
docker-compose build --no-cache
docker-compose up
```

---

## Next Steps

Once all tests pass:
1. ✓ Services run independently
2. ✓ Data persists across restarts
3. ✓ gRPC communication works
4. ✓ Frontend renders correctly
5. Ready for Kubernetes deployment
