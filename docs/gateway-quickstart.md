# API Gateway Quick Start Guide

## Overview

This guide helps you get the Telecom Platform running with Traefik as the API Gateway.

## Prerequisites

- Docker and Docker Compose installed
- Make port 80, 443, and 8080 available
- Domain `api.telecom.com` (or modify configuration)

## Quick Start

### 1. Start the Gateway

```bash
# Make the startup script executable
chmod +x scripts/start-gateway.sh

# Start the platform with API Gateway
./scripts/start-gateway.sh
```

### 2. Configure DNS

Add the domain to your hosts file:

```bash
# On Linux/macOS
echo "127.0.0.1 api.telecom.com" | sudo tee -a /etc/hosts

# On Windows (as Administrator)
echo "127.0.0.1 api.telecom.com" >> C:\Windows\System32\drivers\etc\hosts
```

### 3. Access Services

- **Traefik Dashboard**: http://localhost:8080
- **API Documentation**: https://api.telecom.com/api/v1/swagger
- **Web Dashboard**: http://localhost:3000
- **Health Check**: https://api.telecom.com/api/v1/health

## API Endpoints

All services are now accessible through the gateway:

### API Server
- `https://api.telecom.com/api/v1/subscribers`
- `https://api.telecom.com/api/v1/auth`
- `https://api.telecom.com/api/v1/users`

### Charging Engine
- `https://api.telecom.com/v1/credit/{ip}/check`
- `https://api.telecom.com/v1/credit/{ip}/deduct`
- `https://api.telecom.com/v1/usage`

### Carrier Connector
- `https://api.telecom.com/v1/es2/download`
- `https://api.telecom.com/v1/carrier/profile`

### Packet Gateway
- `https://api.telecom.com/v1/packet/flow`
- `https://api.telecom.com/v1/gateway/status`

## Security Features

### Rate Limiting
- **API endpoints**: 100 requests/minute, burst 200
- **Charging endpoints**: 1000 requests/minute, burst 2000
- **Carrier endpoints**: 50 requests/minute, burst 100

### Security Headers
- HTTPS redirection (80 -> 443)
- XSS protection
- Content Security Policy
- Frame protection

### Authentication
- JWT validation (configure JWT_SECRET env var)
- CORS handling
- Request compression

## Monitoring

### Traefik Dashboard
Access: http://localhost:8080

Shows:
- Active routers and services
- Request metrics
- Health checks
- Middleware status

### Logs
```bash
# View all logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f traefik
docker-compose logs -f api-server
docker-compose logs -f charging-engine
```

### Metrics
Prometheus metrics available at: http://localhost:8080/metrics

## Configuration

### Environment Variables
```bash
# JWT Secret (required for production)
export JWT_SECRET="your-32-character-secret-here"

# Optional: Custom domain
# Update traefik/dynamic/middlewares.yml and service labels
```

### Custom Domain
1. Update service labels in `docker-compose.yml`
2. Update middleware configuration in `traefik/dynamic/middlewares.yml`
3. Update your DNS/hosts file

### SSL Certificates
For production, add certificates to `traefik/certs/` and update configuration:

```yaml
# traefik/traefik.yml
entryPoints:
  websecure:
    address: ":443"
    tls:
      certificates:
        - certFile: /traefik/certs/api.telecom.com.crt
          keyFile: /traefik/certs/api.telecom.com.key
```

## Troubleshooting

### Service Not Accessible
1. Check if service is running: `docker-compose ps`
2. Check service logs: `docker-compose logs [service-name]`
3. Verify Traefik configuration: http://localhost:8080/dashboard/

### Rate Limiting Issues
1. Check middleware configuration in `traefik/dynamic/middlewares.yml`
2. Monitor Traefik dashboard for rate limit status
3. Adjust limits in configuration

### Authentication Issues
1. Verify JWT_SECRET is set: `echo $JWT_SECRET`
2. Check JWT token format and expiration
3. Review Traefik logs for authentication errors

### Performance Issues
1. Monitor Traefik dashboard for request metrics
2. Check service health endpoints
3. Review resource usage: `docker stats`

## Migration from Direct Access

If you were previously accessing services directly:

### Before
```bash
# Direct access
curl http://localhost:8000/api/v1/subscribers
curl http://localhost:8081/v1/health
```

### After
```bash
# Through gateway
curl https://api.telecom.com/api/v1/subscribers
curl https://api.telecom.com/v1/health
```

### Update Client Applications
1. Change base URLs from `http://localhost:PORT` to `https://api.telecom.com`
2. Update authentication headers if needed
3. Handle HTTPS certificate validation

## Production Considerations

1. **Set strong JWT_SECRET**: Use environment variable, not default
2. **Use real SSL certificates**: Obtain from Let's Encrypt or your CA
3. **Monitor resource usage**: Traefik adds minimal overhead
4. **Backup configuration**: Save `traefik/` directory
5. **Log rotation**: Configure Docker logging drivers
6. **Health monitoring**: Set up alerts for service downtime

## Support

- Check logs: `docker-compose logs -f traefik`
- View dashboard: http://localhost:8080
- Review configuration: `traefik/traefik.yml` and `traefik/dynamic/`
- Test endpoints: Use curl or Postman with proper headers
