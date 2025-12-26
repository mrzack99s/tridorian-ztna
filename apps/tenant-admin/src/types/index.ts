export interface User {
    id: string;
    email: string;
    name: string;
    change_password_required: boolean;
}

export interface Tenant {
    id: string;
    name: string;
    slug: string;
    primary_domain: string;
    google_admin_email?: string;
    google_client_id?: string;
}

export interface AccessPolicy {
    id: string;
    name: string;
    effect: 'allow' | 'deny';
    priority: number;
    destination?: string;
    root_node?: PolicyNode;
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
}

export interface Node {
    id: string;
    name: string;
    status: string;
    version: string;
    hostname: string;
    auth_token?: string;
}

export interface Admin {
    id: string;
    name: string;
    email: string;
}
