# Cortex Gateway

![License](https://img.shields.io/github/license/weeco/cortex-gateway.svg?color=blue)
[![Go Report Card](https://goreportcard.com/badge/github.com/weeco/cortex-gateway)](https://goreportcard.com/report/github.com/weeco/cortex-gateway)
![GitHub release](https://img.shields.io/github/release-pre/weeco/cortex-gateway.svg)
[![Docker Repository on Quay](https://quay.io/repository/weeco/cortex-gateway/status "Docker Repository on Quay")](https://quay.io/repository/weeco/cortex-gateway)

Cortex Gateway is a microservice which strives to help you administrating and operating your [Cortex](https://github.com/cortexproject/cortex) Cluster in multi tenant environments.

## Features

- [x] Authentication of Prometheus & Grafana instances with JSON Web Tokens
- [x] Prometheus & Jager instrumentation, compatible with the rest of the Cortex microservices

#### Authentication Feature

If you run Cortex for multiple tenants you need to identify your tenants every time they send metrics or query them. This is needed to ensure that metrics can be ingested and queried separately from each other. For this purpose the Cortex microservices require you to pass a Header called `X-Scope-OrgID`. Unfortunately the Prometheus Remote write API has no config option to send headers and for Grafana you must provision a datasource to do so. Therefore the suggested Cortex k8s manifests suggest to deploy an NGINX cluster inside of each tenant which acts as reverse proxy and does nothing but proxying the traffic and sets the `X-Scope-OrgID` header for your tenant.

We try to solve this problem by adding a Gateway which can be considered the entrypoint for all requests towards Cortex (see [Architecture](#architecture)). Prometheus and Grafana both sent a JSON Web Token (JWT) along with each request. This JWT carries a claim which is the tenant's identifier. Once this JWT is validated we'll set the required `X-Scope-OrgID` header and pipe the traffic to the upstream Cortex microservices (distributor / query frontend).

## Architecture

![Cortex Gateway Architecture](./docs/imgs/architecture.png)

## Configuration

| Flag | Description | Default |
| --- | --- | --- |
| `-gateway.distributor.address` | Upstream HTTP URL for Cortex Distributor | (empty string) |
| `-gateway.query-frontend.address` | Upstream HTTP URL for Cortex Query Frontend | (empty string) |
| `-gateway.auth.jwt-secret` | HMAC secret to sign JSON Web Tokens | (empty string) |

### Expected JWT payload

The expected Bearer token payload can be found here: https://github.com/weeco/cortex-gateway/blob/master/gateway/tenant.go#L7-L11

- "tenant_id"
- "aud"
- "version"

The audience and version claim is currently unused, but might be used in the future (e. g. to invalidate tokens).
