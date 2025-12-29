#!/bin/bash
kubectl rollout restart deploy/gateway-controlplane -n tridorian-ztna
kubectl rollout restart deploy/auth-api -n tridorian-ztna
kubectl rollout restart deploy/management-api -n tridorian-ztna
kubectl rollout restart deploy/tenant-admin -n tridorian-ztna
kubectl rollout restart deploy/backoffice -n tridorian-ztna