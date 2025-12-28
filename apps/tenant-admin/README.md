# Tridorian Tenant Admin

Management portal for individual Tenants/Organizations.

## Features
- **Secure Login**: Local administrator authentication.
- **Setup Wizard**: First-time configuration for:
  - Domain Activation (Free or Custom).
  - Google Identity (OIDC) Integration.
- **Dashboard**: (Coming Soon) Manage policies and users.

## Tech Stack
- React + Vite
- TypeScript
- Vanilla CSS (Premium Design System)
- Vite Proxy for Backend API & Auth

## Getting Started

1.  Navigate to `apps/tenant-admin`
2.  Install dependencies:
    ```bash
    npm install
    ```
3.  Run development server:
    ```bash
    npm run dev
    ```

The app will be available at `http://localhost:5173`. 
**Note**: To access as a specific tenant, ensure you are using the correct hostname (e.g. `[tenant-slug].devztna.rattanaburi.ac.th`) and that it maps to localhost.
