# Multikube Web

Vue 3 frontend for Multikube. Uses Element Plus for the UI and communicates with the gRPC-gateway backend via a Vite dev proxy.

## Setup

```sh
npm install
```

## Development

```sh
npm run dev
```

The dev server proxies `/api/v1` requests to `https://localhost:6443`.

## Production Build

```sh
npm run build
```

## Type Check

```sh
npm run type-check
```
