# Pod Crash Injector

这是一个 Kubernetes 准入 Webhook，用于通过标签控制修改 Pod 的 entrypoint。该工具支持所有 Kubernetes 版本，不依赖于 kubedebug 临时容器功能。

## 功能特点

- 通过标签控制 Pod 的 entrypoint 修改
- 支持自定义 entrypoint 路径
- 支持所有 Kubernetes 版本
- 使用 TLS 加密通信
- 简单易用的部署方式

## 前置要求

- Kubernetes 集群
- kubectl 命令行工具
- openssl（用于生成证书）

## 快速开始

1. 克隆仓库：
```bash
git clone https://github.com/laik/pod-crash-injector.git
cd pod-crash-injector
```

2. 构建容器镜像：
```bash
docker build -t pod-crash-injector:latest .
```

3. 生成证书并部署：
```bash
cd config/cert
chmod +x gen-cert.sh
./gen-cert.sh
```

4. 验证部署：
```bash
kubectl get pods -n pod-crash-injector
```

## 使用方法

要修改 Pod 的 entrypoint，需要给 Pod 添加以下标签：

- `badpod`: 任意值，表示需要修改该 Pod
- `entrypoint`: 要修改的 entrypoint 路径（可选，默认为 `/bin/bash`）

示例：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: example-pod
  namespace: default
  labels:
    badpod: "true"
    entrypoint: "/bin/bash"
spec:
  containers:
  - name: example
    image: nginx
```

## 工作原理

1. Webhook 服务器监听 Pod 的创建和更新事件
2. 当检测到 Pod 具有指定的标签时，修改其 entrypoint
3. 修改后的 Pod 将使用新的 entrypoint 启动

## 配置说明

Webhook 服务器支持以下命令行参数：

- `--port`: Webhook 服务器端口（默认：8443）
- `--tlsCertFile`: TLS 证书文件路径（默认：/etc/webhook/certs/tls.crt）
- `--tlsKeyFile`: TLS 私钥文件路径（默认：/etc/webhook/certs/tls.key）

## 贡献指南

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License 