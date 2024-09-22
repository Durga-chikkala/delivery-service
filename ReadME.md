# Delivery-service
- It serves all the active campaigns matching the targeting rules

### Env's 
```
APP_NAME=delivery_service
HTTP_PORT=5000

LOG_LEVEL=INFO
LOG_TYPE=JSON // supports TEXT, JSON

MONGO_URI=mongodb://localhost:27017
MONGO_DB_NAME=delivery_service

REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=""
REDIS_DB="0"
```

### Technologies Used
- Go: The primary language for building the API.
- MongoDB: Used for data storage of campaigns and targeting rules.
- Redis: Caching layer to speed up data retrieval.

### Testing
- Unit tests are provided in the handlers,services,stores packages. To run the tests, use:

```bash
go test ./...
```
### API Endpoints
#### GET /v1/delivery: 
Retrieve active campaigns based on targeting rules.QueryParam: app, os, country

#### GET /metrics: 
Retrieve Prometheus metrics.

### Metrics
#### Metrics are collected using Prometheus and can be viewed at the /metrics endpoint. This includes:

- Total requests
- Successful responses
- Error rates
- Cache hit and miss rates

