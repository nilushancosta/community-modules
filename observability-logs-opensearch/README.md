# Observability Logs Module for OpenSearch

This module collects logs using Fluent Bit and stores them in OpenSearch.

## Installation

```bash
helm install observability-logs-opensearch \
  oci://ghcr.io/openchoreo/charts/observability-logs-opensearch \
  --create-namespace \
  --namespace openchoreo-observability-plane
```
