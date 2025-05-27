#!/bin/bash

set -e

SERVICE="pod-crash-injector"
NAMESPACE="pod-crash-injector"
SECRET_NAME="webhook-certs"
LOCAL_IP="10.1.201.205"

# 生成 CA 证书
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 365 -key ca.key -subj "/CN=${SERVICE}.${NAMESPACE}.svc" -out ca.crt

# 生成服务器证书
openssl req -newkey rsa:2048 -nodes -keyout server.key -subj "/CN=${SERVICE}.${NAMESPACE}.svc" -out server.csr

# 创建证书配置文件
cat > server.conf << EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${SERVICE}.${NAMESPACE}.svc
DNS.2 = ${SERVICE}.${NAMESPACE}.svc.cluster.local
IP.1 = ${LOCAL_IP}
EOF

# 使用配置文件生成证书
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365 -extensions v3_req -extfile server.conf

# 创建 Kubernetes secret
kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
kubectl create secret tls ${SECRET_NAME} \
    --cert=server.crt \
    --key=server.key \
    --namespace=${NAMESPACE} \
    --dry-run=client -o yaml | kubectl apply -f -

# 获取 CA bundle
CA_BUNDLE=$(cat ca.crt | base64 | tr -d '\n')

# 更新 webhook 配置
sed -e "s|\${CA_BUNDLE}|${CA_BUNDLE}|g" ../webhook-local.yaml | kubectl apply -f -

# 清理临时文件（保留证书文件）
rm -f ca.srl server.csr server.conf 