# PHP Log to Sentry Proxy

This application reads PHP application logs through HTTP with basic authentication, transforms them into Sentry format, and sends them to a Sentry server.

## Features

- Fetches logs from a protected HTTP endpoint with basic authentication
- Parses PHP/Yii2 log format
- Transforms logs into Sentry events
- Sends events to Sentry
- Configurable check interval
- Graceful shutdown handling

## Configuration

Create a `config/config.yaml` file with the following structure:

```yaml
server:
  log_url: "https://your-domain.com/logs"
  username: "your-username"
  password: "your-password"
  check_interval: "5m"  # How often to check for new logs

sentry:
  dsn: "https://your-sentry-dsn"
  environment: "production"
  project: "your-project-name"

logging:
  level: "info"
  file: "app.log"
```

## Building and Running

1. Install dependencies:
```bash
go mod tidy
```

2. Build the application:
```bash
go build -o app-sentry-proxy cmd/app/main.go
```

3. Run the application:
```bash
./app-sentry-proxy
```

## Log Format

The application expects logs in the following format:
```
[2024-03-20 10:00:00] [error] Message here {"key": "value"}
```

Where:
- First part is the timestamp
- Second part is the log level
- Third part is the message
- Optional fourth part is JSON context

## Dependencies

- Go 1.16 or later
- github.com/getsentry/sentry-go
- gopkg.in/yaml.v3 