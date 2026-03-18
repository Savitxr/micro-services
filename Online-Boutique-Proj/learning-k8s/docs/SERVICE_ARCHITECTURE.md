# SERVICE_ARCHITECTURE.md: System Design & Patterns

## Overview
This document explains the high-level architecture, service responsibilities, communication patterns, and failure scenarios.

---

## System Diagram

```
┌──────────────────────────────────────────────────────────────┐
│                   User Browser                               │
└────────────────────────┬─────────────────────────────────────┘
                         │
                    HTTP │ :8080
                         │
┌────────────────────────▼─────────────────────────────────────┐
│                    FRONTEND                                  │
│  (Go) HTTP server, renders HTML                              │
│  ├─ Session management                                       │
│  ├─ HTML templating                                          │
│  └─ gRPC client connections                                  │
└────────┬───────────────────────────────────────┬─────────────┘
         │                                       │
    gRPC │ :3550                            gRPC │ :7070
         │                                       │
┌────────▼─────────────────┐         ┌──────────▼────────────┐
│  PRODUCT CATALOG         │         │   CART SERVICE        │
│  (Go)                    │         │   (C# / .NET)         │
│  - Stateless             │         │   - Stateless         │
│  - Read-only operations  │         │   - State persistence  │
│  - Query products        │         │   - Add/remove items   │
│  - Search/filter         │         │   - Get user cart      │
└────────┬─────────────────┘         └──────────┬────────────┘
         │                                       │
         └───────────────┬───────────────────────┘
                         │
                    SQL  │ :5432
                         │
        ┌────────────────▼─────────────────┐
        │     POSTGRESQL DATABASE          │
        │  - products table                │
        │  - user_carts table              │
        │  - cart_items table              │
        │  - Persistent storage            │
        └──────────────────────────────────┘
```

---

## Service Responsibilities

### Frontend
**Purpose**: User-facing web interface; gRPC client orchestration

**Responsibilities**:
- Accept HTTP requests from browsers
- Render HTML UI with product catalog
- Manage user sessions
- Call ProductCatalogService for product data
- Call CartService for user shopping carts
- Handle form submissions, redirects

**Dependencies**: ProductCatalogService, CartService

**Failure Impact**:
- If Frontend crashes: Users cannot access UI
- If Frontend is slow: User experience degrades
- Frontend is stateless: Easy to replace/restart

---

### ProductCatalogService
**Purpose**: Product data management and queries

**Responsibilities**:
- Serve product listing (ListProducts)
- Fetch individual product (GetProduct)
- Search and filter products (SearchProducts)
- Query PostgreSQL for product data
- Cache most-used products (optional optimization)

**Dependencies**: PostgreSQL database

**Failure Impact**:
- If ProductCatalogService crashes: Users cannot browse products
- If ProductCatalogService is slow: Page loading is slow
- If database is unreachable: Service cannot start
- Service is stateless: Easy to restart; no session loss

**Data Model**:
```sql
products
├── id (primary key)
├── name
├── description
├── picture_url
├── price_usd (currency_code, units, nanos)
└── categories (array)
```

---

### CartService
**Purpose**: User shopping cart management

**Responsibilities**:
- Add items to cart (AddItem)
- Retrieve user cart (GetCart)
- Empty cart (EmptyCart)
- Persist cart state in PostgreSQL
- Manage user-specific data

**Dependencies**: PostgreSQL database

**Failure Impact**:
- If CartService crashes: Users cannot manage carts
- If CartService is slow: Add/remove operations lag
- If database is unreachable: Service cannot start
- Service is stateless: Data persists; new instance resumes operation

**Data Model**:
```sql
user_carts
├── user_id (primary key)
├── created_at
└── updated_at

cart_items
├── id (primary key)
├── user_id (foreign key → user_carts)
├── product_id (foreign key → products)
├── quantity
├── created_at
└── updated_at
```

---

### PostgreSQL Database
**Purpose**: Persistent state storage

**Responsibilities**:
- Durably store product catalog
- Durably store user carts and items
- Enforce data integrity (primary keys, foreign keys)
- Support concurrent queries from services
- Initialize schema on startup

**Failure Impact**:
- If Database crashes: All services lose connectivity
- If Database is offline: ALL operations fail
- **Critical component**: Requires persistent storage, backups, recovery planning

---

## Dependency Graph

```
FRONTEND
├── → ProductCatalogService (gRPC :3550)
│    ├── → PostgreSQL (TCP :5432)
│    └── ✓ No other service dependencies
│
└── → CartService (gRPC :7070)
     └── → PostgreSQL (TCP :5432)

NO cross-backend service dependencies:
- ProductCatalogService does NOT call CartService
- CartService does NOT call ProductCatalogService
- Independent failure isolation
```

---

## Communication Patterns

### Request-Response (RPC)

**Frontend → ProductCatalogService**:
```
Frontend                    ProductCatalogService
  │                                    │
  ├─ (gRPC) ListProducts()            │
  │                                    ├─ Query PostgreSQL
  │                                    │
  │                   (response) ←─────┤
  │ [Product array]                    │
```

**Frontend → CartService**:
```
Frontend                    CartService
  │                                    │
  ├─ (gRPC) AddItem(user_id, product) │
  │                                    ├─ Insert into PostgreSQL
  │                                    │
  │                   (response) ←─────┤
  │ [OK]                               │
```

### Database Queries

Both services query PostgreSQL independently:
- **ProductCatalogService**: `SELECT * FROM products`
- **CartService**: `SELECT * FROM cart_items WHERE user_id = ?`
- No service-to-service calls; direct DB connections

---

## Failure Scenarios

### Scenario 1: ProductCatalogService Crashes

```
State Before:  ✓ Frontend, ProductCatalog, CartService, PostgreSQL
               
After Crash:   ✓ Frontend running, ✗ ProductCatalog down, ✓ CartService, ✓ PostgreSQL

User Impact:   
  - Browse products: FAILS (ProductCatalog unreachable)
  - View/modify cart: SUCCESS (CartService still works)
  
Recovery:      
  1. Detect: Kubernetes liveness probe detects unresponsive service
  2. Replace: Kill pod, start new pod automatically
  3. Resume: Service queries PostgreSQL (data persists)
```

### Scenario 2: CartService Crashes

```
Before:        ✓ Frontend, ProductCatalog, CartService, PostgreSQL

After Crash:   ✓ Frontend, ✓ ProductCatalog, ✗ CartService, ✓ PostgreSQL

User Impact:
  - Browse products: SUCCESS (ProductCatalog works)
  - View/modify cart: FAILS (CartService unreachable)
  - Cart data: SAFE (persisted in PostgreSQL)

Recovery:
  1. Detect: Liveness probe fails
  2. Replace: Auto-restart pod
  3. Resume: Service reconnects to PostgreSQL, data intact
```

### Scenario 3: PostgreSQL Database Crashes

```
Before:        ✓ Frontend, ProductCatalog, CartService, ✓ PostgreSQL

After Crash:   ✓ Frontend, ✗ ProductCatalog (no DB), ✗ CartService (no DB), ✗ PostgreSQL

User Impact:
  - Browse products: FAILS
  - View/modify cart: FAILS
  - ALL operations blocked

Recovery:
  1. Detect: Services cannot connect; health checks fail
  2. Restart: PostgreSQL pod restarts
  3. Mount: Persistent volume remounts (data intact)
  4. Services: Auto-reconnect once DB is healthy
  5. Resume: All operations resume

Time to Recovery: ~1-2 minutes (volume remount + health checks)
```

### Scenario 4: Data Corruption (Rare)

```
Issue:         Product prices corrupted in database

Prevention:    
  - Backups: Daily/weekly snapshots
  - Validation: Schema constraints, foreign keys
  - Testing: Restore procedures tested monthly

Recovery:      
  1. Detect: Monitoring alerts on data anomalies
  2. Backup: Restore from last clean backup
  3. Verify: Validate restored data
  4. Apply: Replay transactions after backup point (if possible)

Time to Recovery: 15-30 minutes depending on backup strategy
```

---

## Scalability Considerations

### Current Configuration (Learning Stage)

```
1 Frontend instance  → Handles moderate traffic
1 ProductCatalog    → Serves read-only data (easily cacheable)
1 CartService       → Handles user-specific data
1 PostgreSQL        → Centralized state
```

### Future Scaling Patterns

**Read-Heavy (Product Catalog)**:
```
Multiple Frontend instances (stateless)
Multiple ProductCatalog instances (stateless, query cache)
  ↓ (all query same PostgreSQL)
PostgreSQL with read replicas
```

**Write-Heavy (Cart Service)**:
```
Multiple Frontend instances
Multiple CartService instances (distributed session affinity)
PostgreSQL with primary-replica setup
  + Cache (Redis) for "user's current cart"
```

**Database Scaling**:
```
PostgreSQL primary (writes)
PostgreSQL read replicas (ProductCatalog queries)
Redis cache (hot product data, user sessions)
```

---

## Resilience Patterns

### Pattern: Stateless Services

**ProductCatalogService & CartService**: Stateless
- No local storage
- No server affinity required
- Failed pod replaced instantly
- Easy horizontal scaling

### Pattern: Health Checks

Both services implement gRPC health checks:
```
Kubernetes liveness probe detects failures
  → Pod terminated
  → New pod auto-started
  → Service restored
```

### Pattern: Graceful Degradation

If CartService unavailable:
- Frontend can still show products
- Shopping workflow incomplete but not blocked

If ProductCatalog unavailable:
- Frontend cannot show products
- No graceful fallback (requires full recovery)

### Pattern: Database Resilience

PostgreSQL Persistent Volume:
- Data survives pod deletion
- PVC auto-remounts on new pod
- RTO: ~2-5 minutes (pod startup + PVC remount)

---

## Anti-Patterns to Avoid

| Pattern | Why It's Bad | Solution |
|---------|-------------|----------|
| Storing state in file system | Pod deletion = data loss | Use database + persistent volumes |
| Service A calls Service B calls Service C | Cascading failures | Minimize service dependencies (current: none!) |
| Hardcoded service IPs | IP changes on restart → failures | Environment variables + service discovery |
| No health checks | Failed services undetected | gRPC health checks + K8s probes |
| Manual restarts | Error-prone, slow recovery | Kubernetes auto-restart policies |

---

## Monitoring & Observability (Future)

Key metrics to track:

| Metric | Service | Target |
|--------|---------|--------|
| Response time | ProductCatalog | < 200ms |
| Response time | CartService | < 100ms |
| Database connection pool | Both | < 80% usage |
| Disk usage | PostgreSQL | < 80% |
| Error rate | All | < 0.1% |
| Pod restarts | All | < 1 per day |

---

## Next Steps

1. **Understand**: Review this architecture with your team
2. **Test**: Use LOCAL_TESTING.md to verify all patterns locally
3. **Simulate**: Use chaos engineering to test failure scenarios
4. **Monitor**: Add Prometheus/Grafana for metrics
5. **Scale**: Practice horizontal scaling with docker-compose replicas
