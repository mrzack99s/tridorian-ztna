# Tridorian ZTNA - Microservices Architecture

## 🏗️ Architecture Overview

The application has been separated into **6 independent services** (4 backend + 2 frontend):

### 1. **Management API** (HTTP REST)
- **Port**: 8080
- **Path**: `cmd/management-api`
- **Dockerfile**: `Dockerfile.management-api`
- **Purpose**: Backoffice & Tenant Administration
- **Responsibilities**:
  - Tenant management (CRUD)
  - Admin user management
  - Policy management (Access & Sign-in)
  - Application management
  - Node management
  - Domain management

### 2. **Gateway Control Plane** (gRPC)
- **Port**: 5443
- **Path**: `cmd/gateway-controlpane`
- **Dockerfile**: `Dockerfile.gateway-controlplane`
- **Purpose**: Gateway orchestration & control
- **Responsibilities**:
  - Gateway registration & authentication
  - Heartbeat monitoring
  - Policy synchronization to gateways
  - Configuration distribution
  - Session IP management

### 3. **Authentication API** (HTTP)
- **Port**: 8081
- **Path**: `cmd/auth-api`
- **Dockerfile**: `Dockerfile.auth-api`
- **Purpose**: User authentication & authorization
- **Responsibilities**:
  - OAuth2/Google authentication
  - JWT token management
  - User session management
  - Gateway listing for VPN users
  - GeoIP-based access control

### 4. **Gateway Agent** (VPN Server)
- **Port**: 6500 (UDP)
- **Path**: `cmd/gateway`
- **Dockerfile**: `Dockerfile.gateway`
- **Purpose**: Edge VPN gateway
- **Responsibilities**:
  - VPN connection handling (QUIC)
  - TUN interface management
  - Firewall policy enforcement
  - User traffic routing
  - Session tracking

### 5. **Tenant Admin** (Frontend)
- **Port**: 3000 (HTTP)
- **Path**: `apps/tenant-admin`
- **Dockerfile**: `Dockerfile.tenant-admin`
- **Purpose**: Web UI for tenant administrators
- **Tech Stack**: React + Vite + Material-UI
- **Features**:
  - Tenant configuration
  - User management
  - Policy configuration
  - Gateway monitoring
  - Application management

### 6. **Backoffice** (Frontend)
- **Port**: 3001 (HTTP)
- **Path**: `apps/backoffice`
- **Dockerfile**: `Dockerfile.backoffice`
- **Purpose**: Web UI for system administrators
- **Tech Stack**: React + Vite + Material-UI
- **Features**:
  - Multi-tenant management
  - System configuration
  - Global monitoring
  - Tenant provisioning

---

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.23+ (for local development)
- PostgreSQL 18
- Valkey (Redis alternative)

### Option 1: Docker Compose (Recommended)

```bash
# Start all backend services
make docker-up

# View logs
make docker-logs-mgmt    # Management API
make docker-logs-cp      # Gateway Control Plane
make docker-logs-auth    # Authentication API

# Stop all services
make docker-down
```

### Option 2: Local Development

**Terminal 1 - Management API:**
```bash
make run-management
```

**Terminal 2 - Gateway Control Plane:**
```bash
make run-controlplane
```

**Terminal 3 - Authentication API:**
```bash
make run-auth
```

**Terminal 4 - Gateway Agent (Optional):**
```bash
NODE_ID=<uuid> CONTROL_PLANE_ADDR=localhost:5443 make run-gateway-agent
```

### Option 3: Build Binaries

```bash
# Build all services
make build-all

# Run individually
./bin/management-api
./bin/gateway-controlplane
./bin/auth-api
./bin/gateway
```

---

## 📋 Available Commands

### Build Commands
```bash
make build-management      # Build Management API
make build-controlplane    # Build Gateway Control Plane
make build-auth           # Build Authentication API
make build-gateway-agent  # Build Gateway Agent
make build-all            # Build all services
```

### Run Commands
```bash
make run-management       # Run Management API (:8080)
make run-controlplane     # Run Control Plane (:5443)
make run-auth            # Run Auth API (:8081)
make run-gateway-agent   # Run Gateway Agent (:6500)
```

### Docker Commands
```bash
make docker-up           # Start all services
make docker-down         # Stop all services
make docker-build        # Build Docker images
make docker-logs-mgmt    # Management API logs
make docker-logs-cp      # Control Plane logs
make docker-logs-auth    # Auth API logs
make docker-logs-gateway # Gateway Agent logs
```

### Utility Commands
```bash
make clean              # Clean build artifacts
make test               # Run tests
make proto              # Generate protobuf code
make install-tools      # Install dev tools
```

---

## 🌐 API Endpoints

### Management API (`:8080`)

#### Tenant Management
- `GET /api/v1/tenants` - List all tenants (Backoffice)
- `POST /api/v1/tenants` - Create tenant
- `DELETE /api/v1/tenants` - Delete tenant
- `GET /api/v1/tenant/me` - Get my tenant
- `PATCH /api/v1/tenant/me` - Update my tenant

#### Admin Management
- `GET /api/v1/admins` - List admins
- `POST /api/v1/admins` - Create admin
- `PATCH /api/v1/admins` - Update admin
- `DELETE /api/v1/admins` - Delete admin

#### Policy Management
- `GET /api/v1/policies/access` - List access policies
- `POST /api/v1/policies/access` - Create access policy
- `PATCH /api/v1/policies/access` - Update access policy
- `DELETE /api/v1/policies/access` - Delete access policy
- `GET /api/v1/policies/sign-in` - List sign-in policies
- `POST /api/v1/policies/sign-in` - Create sign-in policy
- `PATCH /api/v1/policies/sign-in` - Update sign-in policy
- `DELETE /api/v1/policies/sign-in` - Delete sign-in policy

#### Node Management
- `GET /api/v1/nodes` - List nodes
- `POST /api/v1/nodes` - Create node
- `DELETE /api/v1/nodes` - Delete node
- `GET /api/v1/nodes/skus` - List node SKUs
- `GET /api/v1/nodes/sessions` - List active sessions

#### Application Management
- `GET /api/v1/applications` - List applications
- `POST /api/v1/applications` - Create application
- `PATCH /api/v1/applications` - Update application
- `DELETE /api/v1/applications` - Delete application

### Authentication API (`:8081`)

#### Backoffice Authentication
- `POST /auth/backoffice/login` - Backoffice login
- `POST /auth/backoffice/logout` - Backoffice logout
- `GET /auth/backoffice/me` - Get current backoffice user

#### Tenant Admin Authentication
- `POST /auth/mgmt/login` - Tenant admin login
- `POST /auth/mgmt/logout` - Tenant logout
- `GET /auth/mgmt/me` - Get current tenant admin

#### VPN User Authentication
- `GET /auth/` - OAuth2 login (Google)
- `GET /auth/callback` - OAuth2 callback
- `GET /auth/gateways` - List available gateways

### Gateway Control Plane (`:5443` - gRPC)

#### gRPC Methods
- `Register(RegisterRequest)` - Register gateway
- `Heartbeat(HeartbeatRequest)` - Send heartbeat
- `GetConfig(GetConfigRequest)` - Get configuration
- `SyncSessions(SyncSessionsRequest)` - Sync active sessions
- `GetSessionIP(GetSessionIPRequest)` - Assign IP to user

---

## 🔧 Environment Variables

### Management API
```env
APP_ENV=development
MGMT_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=devuser
DB_PASSWORD=P@ssw0rd
DB_NAME=trivpn-trimanaged
CACHE_HOST=localhost
CACHE_PORT=6379
CACHE_PASSWORD=P@ssw0rd
```

### Gateway Control Plane
```env
APP_ENV=development
GRPC_PORT=5443
DB_HOST=localhost
DB_PORT=5432
DB_USER=devuser
DB_PASSWORD=P@ssw0rd
DB_NAME=trivpn-trimanaged
CACHE_HOST=localhost
CACHE_PORT=6379
CACHE_PASSWORD=P@ssw0rd
```

### Authentication API
```env
APP_ENV=development
AUTH_PORT=8081
DB_HOST=localhost
DB_PORT=5432
DB_USER=devuser
DB_PASSWORD=P@ssw0rd
DB_NAME=trivpn-trimanaged
CACHE_HOST=localhost
CACHE_PORT=6379
CACHE_PASSWORD=P@ssw0rd
```

### Gateway Agent
```env
NODE_ID=<uuid>
CONTROL_PLANE_ADDR=localhost:5443
VPN_PORT=6500
HOSTNAME=gateway-1
```

---

## 📦 Docker Images

All services use multi-stage builds for optimal size:

| Service | Dockerfile | Port(s) | Size |
|---------|-----------|---------|------|
| Management API | `Dockerfile.management-api` | 8080 | ~36MB |
| Gateway Control Plane | `Dockerfile.gateway-controlplane` | 5443 | ~33MB |
| Authentication API | `Dockerfile.auth-api` | 8081 | ~35MB |
| Gateway Agent | `Dockerfile.gateway` | 6500/udp | ~33MB |

---

## 🔄 Migration from Monolith

The old `cmd/triztna/main.go` monolith has been split into:

| Old (Monolith) | New (Microservices) |
|----------------|---------------------|
| HTTP API + gRPC in one process | Separated into 4 services |
| Single port (8080 + 5443) | Dedicated ports per service |
| Tight coupling | Loose coupling via gRPC/HTTP |
| Hard to scale | Independent scaling |

**Legacy Dockerfile**: `Dockerfile.tridorian-ztna` (deprecated)

---

## ✅ Benefits of Microservices

1. **Independent Scaling**: Scale each service based on load
2. **Better Resource Allocation**: Optimize resources per service
3. **Fault Isolation**: Issues in one service don't affect others
4. **Easier Deployment**: Deploy services independently
5. **Better Monitoring**: Monitor each service separately
6. **Technology Flexibility**: Use different tech stacks if needed
7. **Team Autonomy**: Different teams can own different services

---

## 🧪 Testing

```bash
# Run all tests
make test

# Test specific service
go test ./cmd/management-api/...
go test ./cmd/gateway-controlpane/...
go test ./cmd/auth-api/...
go test ./cmd/gateway/...
```

---

## 📊 Service Dependencies

```
┌─────────────────┐
│   PostgreSQL    │◄─────┬─────┬─────┬─────┐
└─────────────────┘      │     │     │     │
                         │     │     │     │
┌─────────────────┐      │     │     │     │
│     Valkey      │◄─────┼─────┼─────┼─────┤
└─────────────────┘      │     │     │     │
                         │     │     │     │
                    ┌────▼──┐ ┌▼────┐│    │
                    │ Mgmt  │ │Auth ││    │
                    │ API   │ │API  ││    │
                    └───────┘ └─────┘│    │
                                     │    │
                    ┌────────────────▼┐  ┌▼──────────┐
                    │Gateway Control  │◄─┤  Gateway  │
                    │    Plane        │  │   Agent   │
                    └─────────────────┘  └───────────┘
```

---

## 🔐 Security Notes

- **JWT Authentication**: All APIs use JWT tokens
- **gRPC Auth**: Gateway agents authenticate via tokens
- **Database**: Credentials should be stored in secrets (not hardcoded)
- **TLS**: Production should use TLS for all communications
- **Firewall**: Restrict access to internal services

---

## 📝 Development Workflow

1. **Start Infrastructure**:
   ```bash
   docker-compose up -d postgres valkey
   ```

2. **Run Services Locally**:
   ```bash
   # Terminal 1
   make run-management
   
   # Terminal 2
   make run-controlplane
   
   # Terminal 3
   make run-auth
   ```

3. **Make Changes & Test**:
   ```bash
   make test
   ```

4. **Build & Deploy**:
   ```bash
   make build-all
   make docker-build
   make docker-up
   ```

---

## 🐛 Troubleshooting

### Service won't start
- Check if ports are already in use
- Verify database connection
- Check environment variables

### Database connection failed
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# View PostgreSQL logs
docker-compose logs postgres
```

### gRPC connection refused
- Ensure Gateway Control Plane is running
- Check `CONTROL_PLANE_ADDR` environment variable
- Verify network connectivity

---

## 📚 Additional Resources

- [gRPC Documentation](https://grpc.io/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Valkey Documentation](https://valkey.io/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)

---

## 🤝 Contributing

1. Create a feature branch
2. Make your changes
3. Run tests: `make test`
4. Build all services: `make build-all`
5. Submit a pull request

---

## 📄 License

[Your License Here]
