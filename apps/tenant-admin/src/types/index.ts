export interface User {
    id: string;
    email: string;
    name: string;
    change_password_required: boolean;
    role: 'super_admin' | 'admin' | 'policy_admin';
}

export interface Tenant {
    id: string;
    name: string;
    slug: string;
    primary_domain: string;
    google_admin_email?: string;
    google_client_id?: string;
}


export interface Application {
    id: string;
    name: string;
    description?: string;
}

export interface ApplicationCIDR {
    id: string;
    application_id: string;
    cidr: string;
}

export interface AccessPolicy {
    id: string;
    name: string;
    effect: 'allow' | 'deny';
    priority: number;
    destination_type?: 'cidr' | 'app' | 'sni';
    destination_cidr?: string;
    destination_app_id?: string;
    destination_app?: Application;
    destination_sni?: string;
    root_node?: PolicyNode;
    enabled?: boolean;
    nodes?: Node[];
}

export interface PolicyCondition {
    type: string;
    field: string;
    op: string;
    value: string;
}

export interface PolicyNode {
    id?: string;
    operator: 'AND' | 'OR';
    children?: PolicyNode[];
    condition?: PolicyCondition;
}

export interface SignInPolicy {
    id: string;
    name: string;
    priority: number;
    block: boolean;
    root_node?: PolicyNode;
    enabled?: boolean;
    stage: 'pre_auth' | 'post_auth';
}

export interface Node {
    id: string;
    name: string;
    status: string;
    version?: string; // Kept for backward compat if any
    gateway_version?: string;
    hostname: string;
    ip_address?: string;
    auth_token?: string;
    client_cidr?: string;
    last_seen_at?: string;
    is_active?: boolean;
    node_sku?: NodeSku;
}

export interface NodeSku {
    id: string;
    name: string;
    description?: string;
    max_users: number;
    bandwidth: number;
    price_cents: number;
}

export interface Admin {
    id: string;
    name: string;
    email: string;
    role: 'super_admin' | 'admin' | 'policy_admin';
}
