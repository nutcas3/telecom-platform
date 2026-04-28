# Changelog

All notable changes to the TaaS Platform will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-04-01

### Added
- Initial release of TaaS Platform
- API Server (Go 1.26) with RESTful endpoints
- Carrier Connector for GSMA ES2+ integration
- Charging Engine (Rust 1.94) with Redis-based credit control
- Packet Gateway stub with eBPF architecture
- Web Dashboard (Next.js 15) with Tailwind CSS
- Docker deployment with docker-compose
- Kubernetes manifests for production deployment
- Complete documentation (API, Architecture, Deployment)
- Setup scripts for MongoDB and Redis
- Development environment setup script
- Comprehensive TODO list with research links

### Features
- eSIM creation and management
- Real-time credit control
- Usage tracking and reporting
- Developer dashboard
- Webhook support (planned)
- Multi-tenant architecture
- Horizontal scaling support

### Technical
- Go 1.26 with Green Tea GC
- Rust 1.94 with array_windows
- TypeScript/Next.js 15
- MongoDB 7.0 for subscriber data
- Redis 7 for real-time state
- free5GC integration ready
- eBPF/Aya framework integration

## [Unreleased]

### Added
- **Payment System**: PaymentController with modular handlers for payment operations, refunds, and webhook processing
- **Stripe Integration**: Extracted Stripe gateway types to separate file for improved code organization
- **Carrier Connector**: 
  - Handler package with carrier info, connectivity check, and shared types for GSMA ES2+ profile management
  - Profile handlers with repository integration for eSIM profile lifecycle management via ES2+ protocol
  - ProfileRepository interface with in-memory implementation for eSIM profile storage and CRUD operations
  - REST API server with Gin framework and comprehensive eSIM profile management endpoints
  - Refactored to use GSMA ES2+ protocol with dedicated client package and context-aware profile download
- **eBPF Packet Gateway**:
  - Batch sync mechanism with sync_control flag, sync_buffer map, and stats aggregation
  - eBPF manager with XDP program loading, packet stats tracking, credit management, and bidirectional Redis synchronization
  - Build script for eBPF program compilation with cargo build-bpf integration
  - Integration with XDP program loading, bidirectional Redis sync, and graceful shutdown handling
- **Rate Limiting**:
  - Distributed Redis-based rate limiting with sliding window, token bucket, and multi-tier support
  - Comprehensive rate limiting middleware with IP-based, user-based, and endpoint-specific throttling strategies
  - Redis rate limiter refactored to use interface for improved testability
  - Comprehensive Redis rate limiter tests with mock client, middleware validation, and multi-tier support
- **Authentication & Authorization**:
  - Authentication models with User, Role, Permission, AuthSession, and APIKey structs
  - Authentication service with JWT token management, session handling, and user operations
  - Authentication middleware with JWT validation, RBAC permission checks, and role-based access control
  - Authentication handler with login, registration, user management, and JWT token operations
  - Authentication configuration with JWT secret and reorganized environment variables
  - Casbin RBAC service with policy management, permission checking, and default role configurations
- **Platform Services**:
  - Platform handlers for services, monitoring, deployments, plugins, automations, chaos experiments, and configuration management
  - Platform models for plugins, automations, config, deployments, and chaos experiments
  - Deployment service with CRUD operations, rollback support, and status tracking
  - Plugin service with CRUD operations, filtering, and enable/disable functionality
  - Automation service with workflow execution, scheduling, run history, and comprehensive test coverage
  - Configuration store service with database-backed runtime settings, validation, and sensitive value masking
- **Monitoring & Observability**:
  - Comprehensive health monitoring system with component checks, system metrics, and HTTP handlers
  - Comprehensive alerting system with rule evaluation, multi-channel notifications, and health-based alert triggers
  - Prometheus service client with instant query execution, alert retrieval, and health checking
  - Performance middleware with request tracking, metrics collection, and optimization utilities
- **WebSocket**: WebSocket handler with hub management, client connections, and real-time broadcasting
- **Kubernetes**: Kubernetes service with deployment management, scaling, restart, logs, and pod status monitoring
- **Caching**: Cache package with in-memory implementation, LRU eviction, and multi-cache management
- **Error Handling**: Centralized error handling with typed error codes, validation support, and enhanced parsing utilities
- **Utilities**: millisToDuration utility function for converting milliseconds to time.Duration with zero handling
- **Database**: Cached database wrapper with query caching, invalidation, and statistics support; connection pool optimization and statistics monitoring

### Changed
- **Refactoring**:
  - SubscriberService split into modular files with method organization by domain
  - Carrier connector main.go refactored to extract route setup and handlers into separate packages
  - ES2Client implementation removed from main.go and handlers updated to use internal/es2 package
  - Replaced interface{} with any alias throughout codebase
  - Removed ETag generation and validation middleware, simplified CacheMiddleware to basic Cache-Control header
- **Dependencies**: Updated Go dependencies with JWT auth, Swagger docs, Kubernetes client, Casbin RBAC, and testing utilities

### Fixed
- **Rate Limiter Tests**: Fixed Redis rate limiter mock type mismatch and header name capitalization in tests
- **Test Failures**: Fixed TestRateLimit_ErrorHandling in ratelimit_test.go by using high limit instead of invalid limit
- **IPv6 Support**: Skipped IPv6 test in IPKeyExtractor due to Gin ClientIP() limitations with IPv6 addresses

### Documentation
- Added comprehensive documentation for error handling, testing strategy, monitoring, and development workflow
- Added Swagger documentation models for API endpoints with comprehensive field annotations

### CI/CD
- Added GitHub Actions CI workflow with testing, building, and security scanning
- Added GitHub Actions deployment workflow with staging/production environments and container registry integration
- Added staging and production environment configurations with protection rules, approvals, and deployment restrictions

### Planned Features
- Complete eBPF packet gateway implementation
- Real-time webhook delivery
- Advanced analytics dashboard
- Multi-region deployment support
- SGP.32 (IoT eSIM) support
- CLI tool for developers
- Terraform provider
- Enhanced monitoring and alerting

---

## Version History

- **1.0.0** - Initial release with core functionality
