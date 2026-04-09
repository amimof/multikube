# Design

This document tries to explain the architectural design of Multikube on a high level as well as a pretty low level so you learn how multikube is build fundamentally. Please provide feedback if this document can be improved.  

multikube is built around an API-driven interaction model with loosely coupled API resources. This enables a declarative approach where teams describe *intent* rather than imperative steps, while maintaining a clear separation between infrastructure and configuration responsibilities. This model is what makes Kubernetes so powerful, and Multikube extends the same principles to cluster access and routing bringing modern, API-driven automation to the edge.  

## Components

multikube is delivered as a single statically linked binary. Internally, it is composed of loosely coupled components that are integrated to present a unified and easy-to-manage entry point.

- **Server**: Or the "control plane" which it also is refered to as, is a gRPC server listening on port `5732` by default. This is the component which the `multikubectl` uses to manage state such as creating, updating and deleting *resources*. A resource is a protobuf based API used by administrators to describe intent. When a resource is changed, an event will trigger reconciliation by the controller.
- **Gateway (WIP)**: Provides an HTTP REST-like interface to the Server. It uses the popular `grpc-gateway` to expose gRPC service to web clients such as `curl` or browsers. This components is beeing developed and not yet finished.
- **Controller**: A Go routine that subscribes to events on an internal event bus called `Exchange`. The server emits events, and the controller reacts to them. Its primary responsibility is to compile declarative intent into a runtime configuration used by the proxy to make routing decisions.
- **Proxy**: An HTTP reverse proxy through which all traffic flows. The proxy uses middleware and precompiled runtime configuration to determine how requests should be routed. Compiling configuration ahead of time ensures it can be validated before being applied, keeping the request path fast and predictable.
- **Store**: The persistence layer where all state is stored. Currently, multikube supports a BadgerDB-backed repository or an in-memory store. Work is ongoing to introduce a more scalable storage backend to support larger deployments.
- **CLI**: `multikubectl` is the primary interface for managing one or more multikube instances. It uses a configuration file similar to Kubernetes' kubeconfig, allowing administrators to manage multiple servers through different *contexts*. `multikubectl` communicates with the server over gRPC.
