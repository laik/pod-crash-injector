package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"

	"github.com/laik/pod-crash-injector/pkg/webhook"
	"k8s.io/klog/v2"
)

func main() {
	var parameters webhook.WebhookServer

	// 获取命令行参数
	flag.IntVar(&parameters.Port, "port", 8443, "Webhook server port")
	flag.StringVar(&parameters.CertFile, "tlsCertFile", "/etc/webhook/certs/tls.crt", "File containing the x509 Certificate for HTTPS")
	flag.StringVar(&parameters.KeyFile, "tlsKeyFile", "/etc/webhook/certs/tls.key", "File containing the x509 private key to --tlsCertFile")
	flag.Parse()

	// 创建 Webhook 服务器
	whsvr := webhook.NewWebhookServer()

	// 配置 TLS
	pair, err := tls.LoadX509KeyPair(parameters.CertFile, parameters.KeyFile)
	if err != nil {
		klog.Fatalf("Failed to load key pair: %v", err)
	}

	// 配置 HTTPS 服务器
	server := &http.Server{
		Addr:      fmt.Sprintf(":%v", parameters.Port),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
	}

	// 注册处理函数
	http.HandleFunc("/mutate", whsvr.Serve)

	// 启动服务器
	klog.Info("Starting webhook server...")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		klog.Fatalf("Failed to listen and serve webhook server: %v", err)
	}
}
