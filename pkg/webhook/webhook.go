package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type WebhookServer struct {
	Port     int
	CertFile string
	KeyFile  string
}

func NewWebhookServer() *WebhookServer {
	return &WebhookServer{}
}

func (whsvr *WebhookServer) Serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	klog.Infof("request body: %s", string(body))

	// 验证请求
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	// 解析 AdmissionReview
	admissionReview := admissionv1.AdmissionReview{}
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		klog.Errorf("unmarshal admission review failed: %v", err)
		http.Error(w, "unmarshal admission review failed", http.StatusBadRequest)
		return
	}

	// 处理请求
	admissionResponse := whsvr.handleAdmission(admissionReview)

	// 构造响应
	response := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
	}
	if admissionResponse != nil {
		response.Response = admissionResponse
		if admissionReview.Request != nil {
			response.Response.UID = admissionReview.Request.UID
		}
	}

	// 发送响应
	resp, err := json.Marshal(response)
	if err != nil {
		klog.Errorf("marshal response failed: %v", err)
		http.Error(w, "marshal response failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (whsvr *WebhookServer) handleAdmission(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	req := ar.Request
	if req == nil {
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "empty request",
			},
		}
	}

	// 只处理 Pod 资源
	if req.Kind.Kind != "Pod" {
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	// 解析 Pod
	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		klog.Errorf("unmarshal pod failed: %v", err)
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("unmarshal pod failed: %v", err),
			},
		}
	}

	// 检查 Pod 是否有指定的标签
	if shouldModifyPod(&pod) {
		// 修改 Pod 的 entrypoint
		patch, err := createPatch(&pod)
		if err != nil {
			klog.Errorf("create patch failed: %v", err)
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Message: fmt.Sprintf("create patch failed: %v", err),
				},
			}
		}

		return &admissionv1.AdmissionResponse{
			Allowed: true,
			Patch:   patch,
			PatchType: func() *admissionv1.PatchType {
				pt := admissionv1.PatchTypeJSONPatch
				return &pt
			}(),
		}
	}

	return &admissionv1.AdmissionResponse{
		Allowed: true,
	}
}

func shouldModifyPod(pod *corev1.Pod) bool {
	// 检查 Pod 是否有 badpod 标签
	if pod.Labels == nil {
		return false
	}

	_, hasBadpod := pod.Labels["badpod"]
	if !hasBadpod {
		return false
	}

	// 检查 entrypoint 标签是否存在
	_, hasEntrypoint := pod.Labels["entrypoint"]
	return hasEntrypoint
}

func createPatch(pod *corev1.Pod) ([]byte, error) {
	// 创建修改 entrypoint 的补丁
	patches := []map[string]interface{}{}

	// 获取用户指定的 entrypoint
	entrypoint := pod.Labels["entrypoint"]
	if entrypoint == "" {
		entrypoint = "/bin/bash"
	} else {
		// 处理简单的命令名称
		switch entrypoint {
		case "bash":
			entrypoint = "/bin/bash"
		case "sh":
			entrypoint = "/bin/sh"
		}
	}

	// 为所有容器创建新的命令
	for i := range pod.Spec.Containers {
		// 使用 entrypoint 和 tail -f /dev/null 来保持容器运行
		newCommand := []string{entrypoint, "-c", "tail -f /dev/null"}
		patches = append(patches, map[string]interface{}{
			"op":    "replace",
			"path":  fmt.Sprintf("/spec/containers/%d/command", i),
			"value": newCommand,
		})
	}

	return json.Marshal(patches)
}
