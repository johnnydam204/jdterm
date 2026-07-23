# JDTerm Architecture

## Goals

* Cross-platform (Windows/Ubuntu)
* Go backend + HTML/CSS/JavaScript frontend
* Embedded WebView for desktop release
* Clean Architecture with Dependency Injection

## Layers

Browser(UI) → WebSocket → Router → Handlers → Services (Serial, Config,
Logger...) → OS/Hardware

## Packages

* app: Composition root
* webserver: HTTP/static files
* websocket: Transport only
* router: Message routing
* handlers: Protocol handlers
* serial: UART implementation
* protocol: Public protocol definitions
