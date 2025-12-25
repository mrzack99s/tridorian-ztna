# Tridorian Backoffice

Management portal for Tridorian ZTNA.

## Features
- **Tenant Management**: Create and list organizations.
- **Domain Configuration**: Activate free or custom domains.
- **Identity Provider Setup**: Configure Google Cloud OIDC for each tenant.

## Tech Stack
- React + Vite
- TypeScript
- Material UI (Google Cloud Console-inspired UX)
- Vite Proxy for Backend API

## Getting Started

1.  Navigate to `apps/backoffice`
2.  Install dependencies:
    ```bash
    npm install
    ```
3.  Run development server:
    ```bash
    npm run dev
    ```

The app will be available at `http://localhost:5173`. It expects the Management API to be running on `http://localhost:8080`.
