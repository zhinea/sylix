# Sylix Engine

Sylix Engine is a database management backend focused on Postgres, featuring a control plane and agent architecture. It is designed to provide production-grade Postgres configurations out of the box, including backup storage, Point-in-Time Recovery (PITR), Write-Ahead Logging (WAL), branching, and easy version management.

## 🏗 Architecture

The system consists of three main components:

- **Control Plane**: The master server that manages state, orchestrates operations, and serves the API. It uses SQLite for its internal state.
- **Agent**: Runs on nodes to manage the actual Postgres instances. It connects to the control plane to receive commands.
- **Dashboard**: A modern web interface built with React Router v7 and Vite for managing the system.

## 🚀 Features

- **Postgres Management**: Automated deployment and configuration of Postgres instances.
- **High Availability & Reliability**: Built-in support for WAL and PITR.
- **Backup Management**: Integrated backup storage solutions.
- **Branching**: Support for database branching (similar to Neon).
- **Version Control**: Easy switching between Postgres versions.
- **Modern UI**: A responsive dashboard to monitor and control your database fleet.

## 🛠 Tech Stack

### Backend (Go)
- **Communication**: gRPC (Inter-service) & gRPC-Web (Frontend -> Controlplane).
- **Database**: SQLite (via GORM) for the control plane.
- **Architecture**: Modular monolith with vertical slices (`internal/module`).

### Frontend (TypeScript)
- **Framework**: React Router v7 + Vite.
- **Styling**: Tailwind CSS + shadcn/ui.
- **Network**: Custom gRPC-Web client implementation.

## 🏁 Getting Started

### Prerequisites
- Go 1.22+
- Node.js & pnpm/npm
- Make
- Protoc (for generating protobuf code)

### Running the Backend (Control Plane)

1. **Start the Control Plane**:
   ```bash
   make run
   ```
   The server will start on port `:8082`.

2. **Development Mode** (with hot reload):
   ```bash
   make dev
   ```
   This command also compiles protobufs and builds the agent binary.

### Running the Frontend

1. **Navigate to the dashboard directory**:
   ```bash
   cd ui/dashboard
   ```

2. **Install dependencies**:
   ```bash
   npm install
   ```

3. **Start the development server**:
   ```bash
   npm run dev
   ```
   The dashboard will be available at `http://localhost:5173`.

## 👨‍💻 Development

### Protobuf Generation

If you modify any `.proto` files in the `proto/` directory, you need to regenerate the code for both backend and frontend. The `make dev` command handles this, or you can run them manually:

1. **Backend**:
   ```bash
   make compile-proto
   ```

2. **Frontend**:
   ```bash
   make compile-proto-frontend
   ```

## 📄 License
see [LICENSE](./LICENSE)