# Summary

Redis Prometheus exporter written in Golang. 

The entire solution is packed into Docker, `docker-compose.yaml` starts up the application, `docker-compose.test.yaml` runs exporter tests only, separate `Dockerfile` and `Dockerfile.test` were used for this purpose.
Dockerized app contains Redis, Prometheus, exporter written in Golang. 

Exporter and Prometheus are using OpenSSL self-signed certificates for HTTPS.

## Running the app 

Execute the following commands under `/app` directory:

- Run the application: `docker-compose up --build`
- Run tests: `docker-compose -f .\docker-compose.test.yaml -p ci up --build`

Connect via Prometheus:
- Prometheus is running in the container, it's accessible via `http://localhost:9090/graph` URL in the browser.

Get metrics from HTTPS endpoint:
- `GET` `https://localhost:9999/metrics`

## Project structure overview

The `/app` directory contains `/go` and `/prometheus` subdirectories, `go` contains the exporter written in Golang, Prometheus configuration file is stored in `prometheus` subdirectory.

## Metrics structure
Metrics parsed under INFO generic function are marked with `redis_info` prefix.
Non-numerical values are exposed as labels to `redis_info_non_numerical` metric.  

## Exporter configuration
Settings are stored locally in configuration.yaml file https://github.com/VladimirAndrianov96/exporter/blob/main/app/Go/exporter/cmd/config/configuration.yaml.

Exporter dynamically creates new clients for databases, app will work with either 2 or 5 databases set, required_metrics are also configurable.

## Tests structure
Ginkgo framework and Gomega matcher used for BDD tests.
 
The `gomock` was used to generate interface mocks, mocked Redis client is used for local testing purposes, https://github.com/VladimirAndrianov96/exporter/blob/main/app/Go/exporter/client/mocks/redis_client.go.
