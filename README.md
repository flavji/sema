# SEMA 

SEMA is a real-time, collaborative editing platform designed specifically for building security evaluation reports aligned with the Common Criteria (CC) standard. It offers structured editing, live collaboration, and role-based permissions, streamlining the process for development teams and evaluators.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Dependencies](#dependencies)
- [Setup and Installation](#setup-and-installation)
- [Running the Application](#running-the-application)
- [Running Tests](#running-tests)

---

## Overview

Common Criteria reports are complex, collaborative, and require strict formatting. SEMA is built to simplify that process. This platform provides a web-based editor that supports multi-user collaboration on structured documents, sectioned by the Common Criteria framework.

Documents are broken down into isolated sections and subsections, stored and managed independently. This structure allows granular access control, real-time updates, and better synchronization across large teams working on different parts of a security report.

---

## Features

### Document Management

- Documents are organized by reports, which are further split into sections and subsections.
- Subsections are stored as rich-text deltas (QuillJS), enabling advanced formatting and history tracking.

### Real-Time Collaboration

- Live editing powered by WebSockets.
- Updates are scoped per subsection to avoid edit conflicts.
- Broadcast architecture supports multiple clients editing the same section simultaneously.

### Authentication and Permissions

- Firebase Auth is used for authentication.
- Role-based access:
  - Report Admins: Full access to manage users, sections, and templates.
  - Report Users: Limited to editing specific sections assigned to them.

### Testing and Development Environment

- Integration with Firebase Emulator Suite (Firestore and Auth) for local testing.
- Unit and integration tests written in Go.
- End-to-end test coverage for authentication and document operations.

---

## Architecture

The backend is built using Go and Gin, with modular handlers for authentication, report editing, and WebSocket management. Firestore is used as the primary database to store users, reports, and section data. The frontend is written in basic HTML/CSS/JS with QuillJS for the editor.

---

## Tech Stack

### Backend

- Go (Golang)
- Gin web framework
- Firestore (via Firebase Admin SDK)
- Firebase Auth (with emulator support)
- WebSockets for real-time communication

### Frontend

- HTML / CSS / JavaScript
- QuillJS rich-text editor

### Testing

- Go's testing package
- Firebase Emulator Suite (Firestore, Auth)

---

## Dependencies

### Go Modules

- `cloud.google.com/go/firestore` – Firestore client for database operations  
- `firebase.google.com/go` – Firebase Admin SDK for authentication and Firestore access  
- `github.com/gin-gonic/gin` – Web framework for building REST APIs  
- `github.com/gorilla/websocket` – WebSocket library for real-time communication  
- `github.com/stretchr/testify` – Assertion library for unit tests  
- `google.golang.org/api` – Google APIs required by Firebase SDK  
- `google.golang.org/grpc` – GRPC client library, used by Firestore and Firebase  
- `github.com/dchenk/go-render-quill` – Renders QuillJS deltas to HTML (for previews or tests)  
- `github.com/chromedp/chromedp` – Headless Chrome control for browser-based E2E testing (optional)  
- `github.com/chromedp/cdproto` – Protocol definitions for Chrome DevTools used by chromedp  

### Node.js Packages

- `quill` – Rich text editor for the frontend  
- `quill-delta` – Quill's internal data format library  
- `ws` – WebSocket library for Node.js server and client  
- `nodemon` – Auto-restart for local development (optional)  
- `firebase-tools` – CLI tools for Firebase emulators and deployment  

---

## Setup and Installation

### Prerequisites

- Go 1.20+
- Node.js and Firebase CLI (`npm install -g firebase-tools`)
- Docker (optional for managing services)
- Git

### Clone the Repository

```bash
git clone https://github.com/flavji/sema 
cd sema
```
---

## Running the Application

```bash
go run cmd/app/main.go
```

---

## Running Tests

The project includes unit tests, integration tests with Firebase emulators, and load testing tools.

### Unit Tests

Run unit tests in each component directory:

```bash
cd sema/api/handlers
go test .

cd sema/api/middleware
go test .

cd sema/api/routes
go test .
```

### Integration Tests

#### Firestore Integration


```bash
npm install -g firebase-tools

cd sema/repository
export FIRESTORE_EMULATOR_HOST="localhost:8080"
firebase emulators:start --only firestore &

In another window run... 
go test .
```

#### Firebase Auth Integration

```bash
cd sema/services/authentication
export FIREBASE_AUTH_EMULATOR_HOST="127.0.0.1:9099"
export GCLOUD_PROJECT="test-project"
firebase emulators:start --only auth &

go test .
```

### Other Services

```bash
cd sema/services/reportGeneration
go test .

cd sema/services/websockets
go test .
```

### Load Testing

#### WebSocket Load Testing

In one terminal:

```bash
cd sema/cmd/loadtest
go run .
```

In another terminal:

```bash
cd sema/test/wsloadtest
node ws_loadtest.js
```

#### HTTP Load Testing with k6 (optional)

In one terminal:

```bash
cd sema/cmd/loadtest
go run .
```

In another terminal:


```bash
brew install k6 # or other system package manager equivalent:

cd sema/test/loadtest
k6 run http_test.js
```

---


