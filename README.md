# Flask & GoFeatureFlag Demo

This repository contains a Flask-based application showcasing feature flag integration using **GoFeatureFlag** relay-proxy, along with instructions to integrate the same feature flags into a React frontend. The demo includes:

* **Dark/Light Mode** (`dark_mode`)
* **Promo Banner** (`promo_banner`)
* **Features Grid** (`new_button`, `analytics_section`)
* **Analytics Cards** (`analytics_section`)
* **Chatbot Panel** (`chatbot_enabled`)
* **Footer** (`footer_enabled`)

## Table of Contents

1. [Repository Structure](#repository-structure)
2. [Prerequisites](#prerequisites)
3. [Getting Started](#getting-started)

   * [Clone & Setup](#clone--setup)
   * [Environment Variables](#environment-variables)
   * [Flags Configuration](#flags-configuration)
   * [Relay-Proxy Configuration](#relay-proxy-configuration)
   * [Docker & Docker Compose](#docker--docker-compose)
   * [Run the Demo](#run-the-demo)
4. [Application Overview](#application-overview)
5. [Integrate Flags into React](#integrate-flags-into-react)

   * [Install SDK](#install-sdk)
   * [Initialize Client](#initialize-client)
   * [Use Flags in Components](#use-flags-in-components)
   * [Polling & Cache Behavior](#polling--cache-behavior)
6. [Advanced Topics](#advanced-topics)

   * [Production Deployment](#production-deployment)
   * [Custom Retriever (HTTP, GitHub, etc.)](#custom-retriever-http-github-etc)
7. [Cleaning Up](#cleaning-up)
8. [License](#license)

---

## Repository Structure

```
flask-flag-demo/
├── .env                   # Environment variables (SECRET_KEY)
├── app.py                 # Flask application code
├── requirements.txt       # Python dependencies
├── Dockerfile             # Dockerfile for Flask app
├── docker-compose.yml     # Docker Compose for Flask + relay-proxy
├── flags.yaml             # Local flags configuration
├── goff-proxy.yaml        # Relay-proxy polling and retriever settings
├── README.md              # This file
├── templates/
│   └── index.html         # Flask HTML template
└── static/
    ├── css/
    │   └── style.css      # CSS for light/dark and component styles
    └── js/
        └── script.js      # JS for theme toggling
```

---

## Prerequisites

* **Docker** and **Docker Compose** installed on your machine.
* **Python 3.11+** (if you want to run Flask locally without Docker).
* **Node.js & npm/yarn** (for the React integration).

---

## Getting Started

### Clone & Setup

```bash
git clone https://github.com/your-org/flask-flag-demo.git
cd flask-flag-demo
```

### Environment Variables

1. Create a file named `.env` in the project root:

   ```bash
   ```

touch .env

````
2. Add the Flask secret key:
   ```dotenv
SECRET_KEY=supersecret
````

---

### Flags Configuration

Modify `flags.yaml` to control which features are enabled by default:

```yaml
dark_mode:
  defaultValue: false
promo_banner:
  defaultValue: true
new_button:
  defaultValue: false
analytics_section:
  defaultValue: true
chatbot_enabled:
  defaultValue: true
footer_enabled:
  defaultValue: true
```

* Each key is a feature flag name.
* `defaultValue` can be `true` or `false`.

---

### Relay-Proxy Configuration

Edit `goff-proxy.yaml` if needed. The default polls every 1000 ms (1 second) and uses `flags.yaml`:

```yaml
pollingInterval: 1000
retrievers:
  - kind: file
    path: /goff/flags.yaml
```

---

### Docker & Docker Compose

#### Dockerfile (Flask App)

```dockerfile
FROM python:3.11-slim-bullseye
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE 5000
CMD ["python", "app.py"]
```

#### docker-compose.yml

```yaml
version: '3.8'
services:
  gff-server:
    image: thomaspoignant/go-feature-flag:latest
    entrypoint: ["go-feature-flag"]
    command: ["relay-proxy", "--config", "/goff/goff-proxy.yaml"]
    ports:
      - "8080:1031"
    volumes:
      - ./flags.yaml:/goff/flags.yaml:ro
      - ./goff-proxy.yaml:/goff/goff-proxy.yaml:ro

  flask-app:
    build: .
    ports:
      - "5001:5000"
    depends_on:
      - gff-server
    environment:
      GFF_URL: http://gff-server:1031/client/api/v1
      SECRET_KEY: ${SECRET_KEY}
    volumes:
      - .:/app
```

* The relay proxy listens on container port **1031**, mapped to host **8080**.
* Flask listens on container port **5000**, mapped to host **5001**.

---

### Run the Demo

1. Ensure Docker is running.
2. In the project root:

   ```bash
   ```

docker-compose up --build

````
3. Open `http://localhost:5001` in your browser.
4. Edit `flags.yaml` to toggle features; the relay proxy polls every second, and the Flask endpoint picks changes on page refresh.

---

## Application Overview

- **Flask Server (app.py)**
  - Fetches feature flags from the relay proxy (`GFF_URL`) via HTTP.
  - Renders `index.html`, passing a `flags` dictionary and some sample `stats`.
  - Uses Jinja2 templating to conditionally show/hide sections.

- **Templates & Static Files**
  - `templates/index.html` includes:
    - **Promo Banner** if `flags.promo_banner` is `true`
    - **Header** with Dark/Light toggle
    - **Features Grid** (always-visible card and `new_button`-driven card)
    - **Analytics Section** using a small cards layout (controlled by `analytics_section`)
    - **Chatbot Panel** if `flags.chatbot_enabled` is `true`
    - **Footer** if `flags.footer_enabled` is `true`
  - `static/css/style.css` and `static/js/script.js` control look-and-feel and client-side theme toggling.

- **Relay Proxy (GoFeatureFlag)**
  - Loads `flags.yaml` on startup.
  - Polls the file every second to detect changes, updating its in-memory flag store.
  - Exposes a **client-API** at `/client/api/v1/flags` for the Flask app (and any other client) to fetch all active flags.

---

## Integrate Flags into React

You can share the same relay-proxy–driven flags in a React (or any JavaScript) frontend.

### Install SDK
1. In your React project directory:
   ```bash
npm install @go-feature-flag/js-client-sdk
````

or using Yarn:

```bash
yarn add @go-feature-flag/js-client-sdk
```

---

### Initialize Client

2. In a top-level file (e.g., `src/index.js` or `src/App.js`):

   ```javascript
   ```

import React, { useEffect, useState } from 'react';
import { GoFeatureFlag } from '@go-feature-flag/js-client-sdk';

// Relay proxy base URL (notice '/client/api/v1')
const GFF\_BASE\_URL = '[http://localhost:8080/client/api/v1](http://localhost:8080/client/api/v1)';

const gffClient = GoFeatureFlag({
baseUrl: GFF\_BASE\_URL,
environmentId: '',    // not used for file retriever
pollingInterval: 15000 // milliseconds (15s)
});

function App() {
const \[flags, setFlags] = useState({});

useEffect(() => {
gffClient
.on('ready', (fetchedFlags) => {
setFlags(fetchedFlags);
})
.on('update', (updatedFlags) => {
setFlags(updatedFlags);
});

```
return () => {
  gffClient.close();
};
```

}, \[]);

if (!flags.promo\_banner) {
return <div>Loading...</div>;
}

return (
\<div className={flags.dark\_mode ? 'dark-theme' : 'light-theme'}>
{flags.promo\_banner && <div className="promo">🔥 New Features Available!</div>} <header> <h1>React Dashboard</h1> </header> <main> <section className="features-grid"> <div className="card">Always Available</div>
{flags.new\_button && <div className="card new">New Feature</div>} </section>
{flags.analytics\_section && ( <section className="analytics"> <h2>Analytics Overview</h2>
{/\* ...render your analytics cards... \*/} </section>
)}
{flags.chatbot\_enabled && ( <section className="chatbot-panel"> <h2>Chatbot</h2> <div id="chatbox">Ask me anything!</div> </section>
)} </main>
{flags.footer\_enabled && <footer>© 2025 Demo Corp</footer>} </div>
);
}

export default App;

````
- `GoFeatureFlag({ baseUrl, pollingInterval })` creates a client.
- `on('ready')` is called on initial fetch.
- `on('update')` fires when flags change.
- The client caches flags and polls every `pollingInterval` ms.

---

### Use Flags in Components
3. Anywhere in your React components, read flags from state (or context):
   ```jsx
   if (flags.dark_mode) {
     // apply dark theme class
   }
````

---

### Polling & Cache Behavior

* The JS SDK contacts the relay proxy at:

  * **Startup** (initial fetch)
  * **Every pollingInterval** (e.g., 15s)
* After editing `flags.yaml`, the relay proxy picks up changes within 1s; the JS client sees updates on its next poll.

---

## Advanced Topics

### Production Deployment

* Build Docker images for both:

  * **Relay Proxy** (use `thomaspoignant/go-feature-flag:latest`)
  * **Flask App** (custom image built from `Dockerfile`)
* Deploy behind a load balancer or Kubernetes:

  1. Relay-proxy: reads from a remote store (GitHub, S3, etc.)
  2. Flask and/or React frontend: Nginx serving static React files, and Flask API container.
* Use environment-specific flags by switching retriever configs.

### Custom Retriever (HTTP, GitHub, etc.)

* For remote YAML on GitHub:

  ```yaml
  retrievers:
    - kind: http
      url: https://raw.githubusercontent.com/your-org/your-repo/main/flags.yaml
  pollingInterval: 5000
  ```

---

## Cleaning Up

To stop and remove containers, networks, and volumes:

```bash
docker-compose down --volumes --remove-orphans
```

---

## License

MIT License. See [LICENSE](LICENSE) for details.
