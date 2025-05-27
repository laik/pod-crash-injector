# Pod Crash Injector

这是一个 Kubernetes Mutating Webhook，用于修改 Pod 的 entrypoint，使其使用指定的 shell 并保持容器运行。

## 功能特性

- 通过标签选择器识别需要修改的 Pod
- 支持修改 Pod 的 entrypoint 为指定的 shell
- 自动保持容器运行状态
- 支持本地开发环境

## 本地开发环境设置

### 前置条件

- Go 1.16+
- kubectl 配置正确
- 本地 Kubernetes 集群（如 minikube、kind 等）

### 1. 克隆代码

```bash
git clone https://github.com/laik/pod-crash-injector.git
cd pod-crash-injector
```

### 2. 生成证书

项目提供了自动生成证书的脚本：

```bash
cd config/cert
./gen-cert.sh
```

这个脚本会：
- 生成 CA 证书和服务器证书
- 创建 Kubernetes Secret
- 更新 webhook 配置

### 3. 启动 webhook 服务器

```bash
./scripts/setup-local.sh
go run cmd/webhook/main.go --port=8443 --tlsCertFile=config/cert/server.crt --tlsKeyFile=config/cert/server.key
```

### 4. 测试 webhook

创建一个测试 Pod：

```bash
kubectl run test-pod --image=nginx --labels=badpod=true,entrypoint=bash
```

检查 Pod 状态：

```bash
kubectl get pod test-pod
```

进入容器：

```bash
kubectl exec -it test-pod -- /bin/bash
```

## 配置说明

### Webhook 配置

webhook 配置位于 `config/webhook-local.yaml`，主要配置项：

- `url`: webhook 服务器地址
- `caBundle`: CA 证书
- `rules`: 资源规则
- `failurePolicy`: 失败策略

### 标签说明

Pod 需要添加以下标签才能被 webhook 处理：

- `badpod=true`: 标识需要处理的 Pod
- `entrypoint`: 指定要使用的 shell（支持 `bash`、`sh` 或完整路径）

## 故障排除

1. 证书问题
   - 确保证书正确生成
   - 检查 CA bundle 是否正确配置

2. Webhook 连接问题
   - 确认 webhook 服务器正在运行
   - 检查网络连接和防火墙设置

3. Pod 创建失败
   - 检查 Pod 标签是否正确
   - 查看 webhook 服务器日志

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License 