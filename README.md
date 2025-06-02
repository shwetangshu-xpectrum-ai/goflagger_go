````markdown
# GoFeatureFlag via Docker: Production Deployment and SDK Behavior

This document explains how to use GoFeatureFlag in a Dockerized production environment, including how the relay proxy and Python client SDK interact.

## Deployment Overview: Application and Relay Proxy

In a production environment, two services must run concurrently:

1. **Relay Proxy Container (`go-feature-flag`)**

   - This service is responsible for reading the `flags.yaml` file (or any supported source like S3, GitHub, etc.).
   - It acts as a standalone HTTP microservice, exposing a `/client/api/v1` endpoint.
   - The relay keeps an in-memory cache of all flags and updates it by polling the source file on a configurable interval.
   - All feature evaluation logic happens here — the main application only fetches evaluated values.

2. **Application Container (e.g., Flask App)**

   - This is the main service containing the business logic.
   - It does **not** parse or read the flag file directly.
   - Instead, it makes HTTP requests to the relay proxy’s client API to fetch feature flags.
   - It initializes a GoFeatureFlag Python SDK that interacts with the relay, caching flag values locally.

Both containers must be running, either side by side or with network access between them.

---

## Client-Relay Communication Pattern

The application interacts with the relay proxy to get the latest feature flag values. The process is as follows:

1. **Startup Phase**

   - The application creates an instance of the GoFeatureFlag Python client, setting the `base_url` to the relay proxy’s client API (e.g., `http://gff-server:1031/client/api/v1`).
   - On initialization, the SDK fetches all flags from the relay (`GET /client/api/v1/flags`) and caches them locally.

2. **Runtime Feature Checks**

   - Each time the application checks a flag using `client.bool_variation(...)`, the value is returned from the local in-memory cache.
   - No network request is made during this read.

3. **Automatic Refresh**

   - The SDK includes a background thread that periodically polls the relay to refresh its local flag cache.
   - This ensures that any updates made to the `flags.yaml` file (or remote source) are reflected automatically in the application.

---

## Python SDK Caching and Refresh Logic

The Python SDK is optimized for performance by avoiding repeated network calls during runtime. Here's how it works:

1. **Installation**

   ```bash
   pip install go-feature-flag
````

2. **Initialization Example**

   ```python
   import os
   from go_feature_flag.client import GoFeatureFlagClient

   BASE_URL = os.getenv("GFF_URL", "http://gff-server:1031/client/api/v1")
   refresh_interval = 15  # in seconds

   client = GoFeatureFlagClient(base_url=BASE_URL, refresh_interval_sec=refresh_interval)
   client.wait_for_first_fetch()
   ```

   * `base_url` points to the relay proxy’s endpoint.
   * `refresh_interval_sec` sets how often the SDK will poll the relay for updates.

3. **Using Feature Flags**

   ```python
   is_dark_mode = client.bool_variation("dark_mode", default=False)

   if is_dark_mode:
       # Apply dark mode settings
       ...
   ```

   * The value is retrieved from the local cache.
   * If the cache hasn’t been populated yet, it blocks until the first fetch is complete.

4. **Background Thread Behavior**

   * A background process inside the SDK polls the relay every `refresh_interval_sec` seconds.
   * When the `flags.yaml` file is updated, the relay detects changes based on its own polling interval (configured in `goff-proxy.yaml`).
   * The application picks up those changes during its next polling cycle—no need to restart the app.

5. **Cleanup (Optional)**

   During shutdown, it’s possible to gracefully terminate the background thread:

   ```python
   client._stop_background_thread()
   ```

---

## Folder Structure Example

```
flask-flag-demo/
├── .env
├── app.py              # Flask application using GoFeatureFlagClient
├── requirements.txt
├── Dockerfile
├── docker-compose.yml
├── flags.yaml          # Feature flags definition
├── goff-proxy.yaml     # Relay proxy configuration
└── ...
```

* The `docker-compose.yml` file should define two services:

  * **gff-server**: the relay proxy container.
  * **flask-app**: the main application container.
* The relay proxy can be started with:

  ```yaml
  entrypoint: ["go-feature-flag"]
  command: ["relay-proxy", "--config", "/goff/goff-proxy.yaml"]
  ```

---

## Operational Behavior Summary

* The application starts with a local in-memory cache of flags, fetched from the relay proxy.
* During runtime, all flag checks are handled via this local cache, resulting in fast reads with zero network latency.
* Background threads on both the relay and the SDK ensure that flag updates are reflected without restarting any services.
* This architecture allows for scalable, real-time flag updates with minimal overhead.

---

## Important

* Two containers are essential in production: the application and the relay proxy.
* The relay proxy reads and manages the flags, while the application focuses on consuming them.
* The Python SDK is optimized to minimize traffic, enhance performance, and handle updates automatically.

This setup ensures a reliable, scalable, and efficient way to manage feature flags in production environments.

```
