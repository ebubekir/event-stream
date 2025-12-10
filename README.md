# Event Stream

A backend service that collects event data via HTTP API, stores it in ClickHouse, and provides aggregated metrics.

## What It Does

- Accepts event data (event name, user id, timestamp, device info, etc.)
- Validates and stores events
- Handles duplicate events (idempotency)
- Returns aggregated metrics filtered by time, event name, or platform

## How to Run

### 1. Start ClickHouse

```bash
docker run -d \
  --name clickhouse \
  -p 8123:8123 \
  -p 9000:9000 \
  -e CLICKHOUSE_USER=user \
  -e CLICKHOUSE_PASSWORD=1234 \
  clickhouse/clickhouse-server:latest
```

### 2. Configure the App

Edit `config/config.yaml`:

```yaml
environment_type: "dev"
port: 8080
database_type: "clickhouse"
clickhouse_url: "http://user:1234@localhost:8123"
```

### 3. Run the App

**Option A: With Docker**

```bash
# Build
docker build -t event-stream .

# Run
docker run -d \
  --name event-stream \
  -p 8080:8080 \
  --link clickhouse:clickhouse \
  event-stream
```

**Option B: Without Docker**

```bash
go run ./cmd/api
```

## Project Structure

This project uses **DDD (Domain-Driven Design)** and **Hexagonal Architecture**.

```
internal/
├── domain/        # Business logic (no external deps)
├── application/   # Use cases and services
└── adapter/       # HTTP handlers, DB implementations
```

### Why This Architecture?

- **Easy to test**: Business logic is separate from database and HTTP code
- **Easy to change**: Want to switch from ClickHouse to PostgreSQL? Just change the adapter
- **Clear boundaries**: Each layer has one job

Dependencies flow inward: `adapter → application → domain`

## Why ClickHouse?

This service handles lots of events per second. ClickHouse is built for this:

- **Fast writes**: Can handle millions of rows per second
- **Fast reads**: Column-based storage makes aggregation queries very fast
- **Built for analytics**: Perfect for "count events by name" or "events in last hour" queries
- **Efficient storage**: Compresses data well, saves disk space

## API

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/events` | Create single event |
| POST | `/events/batch` | Create multiple events |
| GET | `/events/metrics` | Get aggregated metrics |

### Swagger UI

API documentation is available at:

```
http://localhost:8080/swagger/index.html
```

### Event Schema

```json
{
  "name": "string (required)",
  "channel_type": "web | mobile | desktop | tv | console | other (required)",
  "timestamp": 1234567890,
  "previous_timestamp": 1234567880,
  "date": "2024-01-15T10:30:00Z",
  "user_id": "user-123",
  "user_pseudo_id": "pseudo-456",
  "event_params": [
    {
      "key": "page_title",
      "string_value": "Home Page"
    },
    {
      "key": "scroll_depth",
      "number_value": 75.5
    },
    {
      "key": "is_logged_in",
      "boolean_value": true
    }
  ],
  "user_params": [
    {
      "key": "subscription",
      "string_value": "premium"
    }
  ],
  "device": {
    "category": "mobile",
    "mobile_brand_name": "Apple",
    "mobile_model_name": "iPhone 14",
    "operating_system": "iOS",
    "operating_system_version": "17.0",
    "language": "en-US",
    "browser_name": "Safari",
    "browser_version": "17.0",
    "hostname": "example.com"
  },
  "app_info": {
    "id": "com.example.app",
    "version": "2.1.0"
  },
  "items": [
    {
      "id": "SKU-001",
      "name": "Blue T-Shirt",
      "brand": "Nike",
      "variant": "Large",
      "price_in_usd": 29.99,
      "quantity": 2,
      "revenue_in_usd": 59.98
    }
  ]
}
```

### Examples

**Create Single Event**

```bash
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{
    "name": "page_view",
    "channel_type": "web",
    "user_id": "user-123",
    "device": {
      "operating_system": "Windows",
      "browser_name": "Chrome"
    },
    "event_params": [
      {"key": "page_title", "string_value": "Home"}
    ]
  }'
```

Response:
```json
{"id": "550e8400-e29b-41d4-a716-446655440000"}
```

**Create Batch Events**

```bash
curl -X POST http://localhost:8080/events/batch \
  -H "Content-Type: application/json" \
  -d '{
    "events": [
      {
        "name": "page_view",
        "channel_type": "web",
        "user_id": "user-123"
      },
      {
        "name": "button_click",
        "channel_type": "mobile",
        "user_id": "user-456"
      }
    ]
  }'
```

Response:
```json
{"ids": ["id-1", "id-2"]}
```

**Get Metrics**

```bash
# Basic metrics
curl "http://localhost:8080/events/metrics?event_name=page_view"

# With time range
curl "http://localhost:8080/events/metrics?event_name=page_view&from=2024-01-01T00:00:00Z&to=2024-01-31T23:59:59Z"

# Group by channel
curl "http://localhost:8080/events/metrics?event_name=page_view&group_by=channel"

# Group by day
curl "http://localhost:8080/events/metrics?event_name=page_view&group_by=daily"

# Group by hour
curl "http://localhost:8080/events/metrics?event_name=page_view&group_by=hourly"
```

Response:
```json
{
  "event_name": "page_view",
  "from": "2024-01-01T00:00:00Z",
  "to": "2024-01-31T23:59:59Z",
  "total_count": 15420,
  "unique_user_count": 3200,
  "grouped_metrics": [
    {"group_key": "web", "total_count": 10000, "unique_user_count": 2000},
    {"group_key": "mobile", "total_count": 5420, "unique_user_count": 1200}
  ]
}
```

## Future Improvements

- [ ] Add Kafka for async event processing
- [ ] Add Redis cache for frequent queries
- [ ] Add rate limiting
- [ ] Add authentication
- [ ] Add Docker Compose setup
- [ ] Add Prometheus metrics
- [ ] Add more aggregation options (by device, by country, etc.)

