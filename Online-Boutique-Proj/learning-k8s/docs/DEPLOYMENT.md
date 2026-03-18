# DEPLOYMENT.md: Docker Image Building & Containerization

## Overview
This document details how each microservice is containerized, built, and optimized for production deployment.

---

## Service-Specific Configurations

### ProductCatalogService (Go + PostgreSQL)

**Dockerfile Strategy**: Multi-stage build for minimal image size

```dockerfile
# Stage 1: Build
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o server .

# Stage 2: Runtime
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/server /app/
EXPOSE 3550
ENTRYPOINT ["/app/server"]
```

**Build Command**:
```bash
docker build -t boutique/product-catalog:v1.0 ./services/product-catalog
```

**Image Size**: ~15-20 MB (optimized with compression flags)

**Optimization Techniques**:
- Multi-stage build: Removes build dependencies from final image
- Alpine base: Minimal footprint (5 MB base)
- Static binary: `CGO_ENABLED=0` for portable Go binary
- Link-time optimization: `-ldflags="-s -w"` removes symbols/debug info

---

### CartService (C# / .NET + PostgreSQL)

**Dockerfile Strategy**: Multi-stage with .NET SDK and runtime separation

```dockerfile
# Stage 1: Build
FROM mcr.microsoft.com/dotnet/sdk:7.0 AS builder
WORKDIR /app
COPY . .
RUN dotnet restore
RUN dotnet publish -c Release -o /app/out

# Stage 2: Runtime
FROM mcr.microsoft.com/dotnet/aspnet:7.0
WORKDIR /app
COPY --from=builder /app/out .
EXPOSE 7070
ENTRYPOINT ["dotnet", "CartService.dll"]
```

**Build Command**:
```bash
docker build -t boutique/cart-service:v1.0 ./services/cart
```

**Image Size**: ~300-400 MB (industry standard for .NET)

**Important .NET Configuration**:
- SDK image (~2GB) used only for compilation
- ASP.NET runtime image (~300MB) for execution
- Build artifacts copied from builder stage
- Environment variables set in docker-compose.yml (not Dockerfile)

---

### Frontend (Go)

**Dockerfile Strategy**: Multi-stage with static file embedding

```dockerfile
# Stage 1: Build
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o frontend .

# Stage 2: Runtime
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/frontend /app/
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/static /app/static
EXPOSE 8080
ENTRYPOINT ["/app/frontend"]
```

**Build Command**:
```bash
docker build -t boutique/frontend:v1.0 ./frontend
```

**Image Size**: ~80-100 MB (includes static assets)

**Special Considerations**:
- Static assets (HTML, CSS, JS) copied from source
- Template files embedded at runtime
- No changes to Frontend code: build as-is

---

## Database Container (PostgreSQL)

**Image**: `postgres:14-alpine`

**Size**: ~80 MB (alpine-based)

**Initialization**:
- SQL scripts executed on first run via `/docker-entrypoint-initdb.d/`
- Schema created by `init.sql`
- Data seeded by `seed-data.sql`
- File order matters: numbered prefixes (01-, 02-, etc.) indicate execution order

**Volume Mounting**:
```yaml
volumes:
  - ./database/init.sql:/docker-entrypoint-initdb.d/01-init.sql
  - ./database/seed-data.sql:/docker-entrypoint-initdb.d/02-seed-data.sql
  - postgres_data:/var/lib/postgresql/data
```

---

## Building for Kubernetes

### Image Registry Setup

1. **Local Development** (docker-compose):
   ```bash
   docker-compose build
   ```

2. **Push to Registry** (e.g., Docker Hub, GCR):
   ```bash
   docker tag boutique/product-catalog:v1.0 gcr.io/my-project/product-catalog:v1.0
   docker push gcr.io/my-project/product-catalog:v1.0
   ```

3. **Kubernetes Reference** (future manifests):
   ```yaml
   containers:
   - name: product-catalog
     image: gcr.io/my-project/product-catalog:v1.0
     imagePullPolicy: IfNotPresent
   ```

### Version Tagging Strategy

- **Development**: `v1.0-dev`
- **Testing**: `v1.0-rc1`
- **Production**: `v1.0`, `latest`

---

## Best Practices Applied

| Practice | Implementation |
|----------|-----------------|
| Minimal Base Images | Alpine for Go, official .NET runtime for C# |
| Multi-Stage Builds | Separate build and runtime stages |
| No Root User | Default container user (alpine, dotnet) |
| Health Checks | Configured at docker-compose level |
| Port Exposure | Each Dockerfile declares EXPOSE |
| Environment Variables | Set in docker-compose/K8s manifests, not in Dockerfile |

---

## Troubleshooting Common Build Issues

**Issue**: "base image not found"
- Solution: Run `docker pull postgres:14-alpine; docker pull golang:1.21-alpine; docker pull mcr.microsoft.com/dotnet/sdk:7.0`

**Issue**: "cannot find package"
- Solution: Ensure go.mod/go.sum are in service directories
- Solution: Ensure *.csproj files are in CartService directory

**Issue**: Build hangs during `dotnet restore`
- Solution: Check network connectivity inside container; add `--progress=plain` for verbose output

---

## Performance Tips

- **Layer Caching**: Docker caches layers; put frequently-changing files (source code) late in Dockerfile
- **Build Context**: Ensure `.dockerignore` excludes node_modules, vendor, build artifacts
- **Parallel Builds**: `docker-compose build --parallel` speeds up multi-service builds

---

## Next Steps

Once comfortable with docker-compose builds, progress to:
1. Kubernetes deployment with Docker registry
2. Container registry (GCR, ECR, Docker Hub)
3. Image scanning for vulnerabilities
4. Multi-architecture builds (amd64, arm64)
