#!/bin/bash

set -e

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# 切换到项目根目录
cd "$PROJECT_ROOT"

# 生成证书
cd config/cert
./gen-cert.sh

# 等待证书生成完成
sleep 2

# 获取 CA bundle
CA_BUNDLE=$(cat ca.crt | base64 | tr -d '\n')

# 更新本地 webhook 配置
cd "$PROJECT_ROOT"
sed -e "s|\${CA_BUNDLE}|${CA_BUNDLE}|g" config/webhook-local.yaml | kubectl apply -f -

# 删除集群内的 webhook 配置（如果存在）
kubectl delete mutatingwebhookconfiguration pod-crash-injector --ignore-not-found=true

echo "本地开发环境设置完成！"
echo "现在你可以运行 webhook 服务器："
echo "cd $PROJECT_ROOT && go run cmd/webhook/main.go --port=8443 --tlsCertFile=config/cert/server.crt --tlsKeyFile=config/cert/server.key"

kubectl delete pod test-pod --ignore-not-found=true && kubectl run test-pod --image=nginx --labels=badpod=true,entrypoint=bash 