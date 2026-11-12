package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// writeJSON 统一输出 JSON：HTTP 状态码和业务 code 是两个东西，这里分开传
func writeJSON(w http.ResponseWriter, status int, code int, message string, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code":    code,
		"message": message,
		"data":    data,
	})
}

func main() {
	mux := http.NewServeMux()

	// 健康检查：确认服务活着
	mux.HandleFunc("/api/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, 0, "ok", map[string]any{"pong": true})
	})

	// 一个带查询参数的接口：/api/v1/hello?name=陶文鑫
	mux.HandleFunc("/api/v1/hello", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "world"
		}
		writeJSON(w, http.StatusOK, 0, "ok", map[string]any{
			"msg": fmt.Sprintf("hello, %s", name),
		})
	})

	addr := ":8090"
	log.Printf("服务启动: http://127.0.0.1%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
