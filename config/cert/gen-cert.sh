#!/bin/bash

set -e

SERVICE="pod-crash-injector"
NAMESPACE="pod-crash-injector"
SECRET_NAME="webhook-certs"

# 生成 CA 证书
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 365 -key ca.key -subj "/CN=${SERVICE}.${NAMESPACE}.svc" -out ca.crt

# 生成服务器证书
openssl req -newkey rsa:2048 -nodes -keyout server.key -subj "/CN=${SERVICE}.${NAMESPACE}.svc" -out server.csr
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365

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
sed -e "s|\${CA_BUNDLE}|${CA_BUNDLE}|g" ../webhook.yaml | kubectl apply -f -

# 清理临时文件
rm -f ca.key ca.crt ca.srl server.key server.csr server.crt 