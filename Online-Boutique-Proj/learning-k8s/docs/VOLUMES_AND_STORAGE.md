# VOLUMES_AND_STORAGE.md: PostgreSQL Persistence & Kubernetes PVCs

## Overview
PostgreSQL is the only stateful component. This document explains data persistence strategy, volume configuration, seeding, and disaster recovery for both docker-compose and Kubernetes.

---

## Docker Compose Storage

### Named Volume Strategy

```yaml
volumes:
  postgres_data:  # Named volume - persists between restarts

services:
  postgres:
    volumes:
      - postgres_data:/var/lib/postgresql/data  # Mount at database data directory
      - ./database/init.sql:/docker-entrypoint-initdb.d/01-init.sql
      - ./database/seed-data.sql:/docker-entrypoint-initdb.d/02-seed-data.sql
```

### Data Persistence Behavior

**On First Run**:
1. Docker creates `postgres_data` volume
2. PostgreSQL initializes fresh database
3. `init.sql` executes → schema created
4. `seed-data.sql` executes → products populated
5. Data stored in `postgres_data` volume

**On Restart** (`docker-compose restart`):
1. PostgreSQL container restarts
2. Existing `postgres_data` volume mounted
3. All data persists (no re-initialization)
4. Schema and products still available

**On Full Cleanup** (`docker-compose down -v`):
1. All containers stopped
2. Network deleted
3. Volumes DELETED (`-v` flag)
4. Next `docker-compose up` restarts fresh

### Volume Management Commands

```bash
# List volumes
docker volume ls

# Inspect volume location
docker volume inspect learning-k8s-postgres_data

# Backup volume data
docker run --rm -v postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres-backup.tar.gz /data

# Restore from backup
docker volume create postgres_data_restored
docker run --rm -v postgres_data_restored:/data -v $(pwd):/backup alpine tar xzf /backup/postgres-backup.tar.gz -C /data

# Remove volume (DANGEROUS - deletes data)
docker volume rm postgres_data
```

---

## Initialization & Seeding

### How It Works

PostgreSQL container executes scripts in `/docker-entrypoint-initdb.d/` on first run:

```dockerfile
# PostgreSQL container behavior
On startup:
  IF database doesn't exist:
    1. Create database
    2. Execute all scripts in /docker-entrypoint-initdb.d/ (sorted alphabetically)
    3. Mark initialization complete
  ELSE (database exists):
    Skip initialization, start normally
```

### Script Execution Order

```
learned-k8s/database/
├── init.sql        → 01-* (schema creation)
└── seed-data.sql   → 02-* (data population)
```

File naming ensures correct order:
- `01-init.sql` runs first (CREATE TABLE statements)
- `02-seed-data.sql` runs second (INSERT statements)

### Idempotent Schema Design

All schema creation uses `IF NOT EXISTS`:

```sql
CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    ...
);

CREATE INDEX IF NOT EXISTS idx_cart_items_user_id ON cart_items(user_id);
```

This allows:
- Safe re-runs without errors
- Safe volume reuse if scripts change

---

## Kubernetes Persistent Volumes (PVC)

### Architecture

```
┌──────────────────────────────────────────────┐
│ Kubernetes Cluster                           │
├──────────────────────────────────────────────┤
│                                              │
│  PostgreSQL Pod (ephemeral/temporary)        │
│  ├─ Volume Mount: /var/lib/postgresql/data  │
│  │                                           │
│  └────────────────┬─────────────────────────┤
│                   │                          │
│  PersistentVolumeClaim (persistent data)    │
│  ├─ StorageClass: fast / standard           │
│  ├─ Size: 10Gi                              │
│  │                                           │
│  └────────────────┬─────────────────────────┤
│                   │                          │
│  PersistentVolume (backing storage)         │
│  ├─ AWS EBS / GCP PD / Azure Disk           │
│  └─ Survives pod deletion                   │
│                                              │
└──────────────────────────────────────────────┘
```

### Kubernetes Manifest (Future Reference)

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: postgres-pv
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteOnce
  storageClassName: fast
  gcePersistentDisk:
    pdName: postgres-disk
    fsType: ext4
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pvc
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: fast
  resources:
    requests:
      storage: 10Gi
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: postgres-init-scripts
data:
  01-init.sql: |
    CREATE TABLE IF NOT EXISTS products (...)
  02-seed-data.sql: |
    INSERT INTO products VALUES (...)
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
spec:
  serviceName: postgres-service
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:14-alpine
        ports:
        - containerPort: 5432
        env:
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: postgres-credentials
              key: username
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-credentials
              key: password
        - name: POSTGRES_DB
          value: boutique_db
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
          subPath: postgres  # Prevent init script conflicts
        - name: init-scripts
          mountPath: /docker-entrypoint-initdb.d
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
      - name: init-scripts
        configMap:
          name: postgres-init-scripts
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: fast
      resources:
        requests:
          storage: 10Gi
```

---

## Storage Classes (Kubernetes)

| StorageClass | Performance | Cost | Use Case |
|--------------|-------------|------|----------|
| `fast` | SSD, high IOPS | $$$ | Production databases |
| `standard` | HDD, general purpose | $ | Dev/test, backups |
| `snapshots` | Local SSD cache | $$ | High-traffic read replicas |

---

## Data Recovery Procedures

### Docker Compose - Backup

```bash
# Create backup directory
mkdir -p backups

# Dump database
docker exec learning-k8s-postgres pg_dump -U postgres boutique_db > backups/boutique_db.sql

# Or backup entire volume
docker run --rm -v postgres_data:/data -v $(pwd)/backups:/backup alpine \
  tar czf /backup/postgres-volume.tar.gz /data
```

### Docker Compose - Restore

```bash
# From SQL dump
cat backups/boutique_db.sql | docker exec -i learning-k8s-postgres psql -U postgres -d boutique_db

# From volume backup
docker volume create postgres_data_restored
docker run --rm -v postgres_data_restored:/data -v $(pwd)/backups:/backup alpine \
  tar xzf /backup/postgres-volume.tar.gz -C /data
```

### Kubernetes - Backup

```bash
# Using kubectl port-forward
kubectl port-forward svc/postgres-service 5432:5432 &
pg_dump -h localhost -U postgres -d boutique_db > backups/k8s-backup.sql

# Or using kubectl exec
kubectl exec -i postgres-0 -- pg_dump -U postgres boutique_db > backups/k8s-backup.sql
```

### Kubernetes - Restore

```bash
# Port-forward method
kubectl port-forward svc/postgres-service 5432:5432 &
psql -h localhost -U postgres -d boutique_db < backups/k8s-backup.sql

# Or using kubectl exec
cat backups/k8s-backup.sql | kubectl exec -i postgres-0 -- psql -U postgres -d boutique_db
```

---

## Disaster Recovery Checklist

| Scenario | Recovery Time | Impact |
|----------|---------------|--------|
| Pod crashes | ~1-2 min | Services briefly unavailable; data intact (PVC survives) |
| Node crashes | ~5-10 min | Pod reschedules; PVC remounts on healthy node |
| Storage corrupted | ~15-30 min | Restore from backup; some data loss possible |
| Accidental deletion | Variable | Restore from backup + snapshot |

**Prevention**:
- Regular backups (daily/weekly)
- Test restore procedures monthly
- Monitor disk space (PVC filling up stops writes)

---

## Performance Considerations

### Database Performance Tuning

```sql
-- Optimize for product queries (mostly reads)
CREATE INDEX idx_products_categories ON products USING GIN(categories);

-- Optimize cart queries (user_id frequent filter)
CREATE INDEX idx_cart_items_user_id ON cart_items(user_id);
CREATE INDEX idx_cart_items_product_id ON cart_items(product_id);
```

### Connection Pooling

**ProductCatalogService** (Go):
```go
// Connection pool in db.go
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(5)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
```

**CartService** (.NET):
```csharp
// Entity Framework connection pool (automatic)
// Default: MaxPoolSize=100
```

### Volume Performance (Kubernetes)

- **SSD (~20-30ms latency)**: Suitable for transactional workloads
- **HDD (~5-10ms latency)**: Adequate for read-heavy catalogs
- **Network storage (EBS, PD)**: Varies; typically 10-50ms

---

## Scaling Considerations

**Read Replicas** (Future):
- PostgreSQL read replicas on same or different node
- Product queries can offload to replicas
- Update primary only

**Sharding** (Rare for this workload):
- Separate database per customer/region
- Complex; avoid unless necessary

**Caching** (Recommended):
- Redis cache product catalog (read-heavy)
- Reduce database queries

---

## Migration: Docker Compose → Kubernetes + Managed Database

Most enterprises migrate to managed databases (Cloud SQL, RDS, Aurora):

1. **Export Data**:
   ```bash
   pg_dump -h localhost boutique_db > export.sql
   ```

2. **Create Managed Database** (Cloud SQL, RDS, etc.)

3. **Import Data**:
   ```bash
   psql -h cloudsql-host -U postgres -d boutique_db < export.sql
   ```

4. **Update Connection String**:
   ```yaml
   env:
   - name: DB_HOST
     value: cloudsql.database.connection  # Managed DB endpoint
   ```

5. **Remove local PostgreSQL** from K8s (no local state needed)

---

## Next Steps

1. **Practice**: Create backups, test restores
2. **Monitor**: Watch disk usage, query performance
3. **Automate**: Cron jobs for regular backups
4. **Scale**: Add read replicas, implement caching
