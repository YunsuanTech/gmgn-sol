package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func main() {
	// 初始化 HTTP 客户端
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_120),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		log.Println("Failed to create HTTP client:", err)
		return
	}

	// 创建每十分钟触发一次的定时器
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// 无限循环，等待定时器触发
	for {
		select {
		case <-ticker.C:
			fetchAndSaveData(client)
		}
	}
}

// fetchAndSaveData 执行数据爬取并保存为 JSON 文件
func fetchAndSaveData(client tls_client.HttpClient) {
	// 创建 HTTP 请求
	req, err := http.NewRequest(http.MethodGet, "https://gmgn.ai/defi/quotation/v1/rank/sol/pump_ranks/1h", nil)
	if err != nil {
		log.Println("Failed to create request:", err)
		return
	}

	// 设置请求头
	req.Header = http.Header{
		"cookie":     {"cf_clearance=9aYSJi1nYB46ZLW37KvvVevlkLmDvlzce5XGwPZE9VM-1736822574-1.2.1.1-3h4a52VprdVeakKe3tZQY6sHimRoxD20JYaCtqSYOmsqiq_q542pOnPFe2RHuFB1CZoCYhV7dhnnO8E5jeDAoEHczmPtUtRuoYKaJ31w490UxscOHm.DeFHC0CJDt9s5t58ul4AyfhwRkSa4Lgm0GbkrocwgeVP4xf1kvpWt2_KYb5VvZIrZnr4fMMPSk6eQKKvfpkfOcYuC219XtH87cLRtS9Y5CeEvXPID5Fw0JsQ7doGC51zaayHpO1fbmH50caON0Fo9BqAMhALQQc0yW56LoF.w_wM781RnmtOrqrdSlVB.MtyPzvsmM2V4CqMC1hjLlcrjQoUoKiFiQF0aZw"},
		"user-agent": {"Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Mobile Safari/537.36"},
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Request failed:", err)
		return
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		log.Printf("Received non-200 status code: %d", resp.StatusCode)
		return
	}

	// 读取响应数据
	readBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read response body:", err)
		return
	}

	// 解析 JSON 数据
	var data interface{}
	if err := json.Unmarshal(readBytes, &data); err != nil {
		log.Println("Failed to unmarshal JSON:", err)
		return
	}

	// 生成文件名
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("data/gmgn_sol_%d.json", timestamp)

	// 创建并写入 JSON 文件
	file, err := os.Create(filename)
	if err != nil {
		log.Println("Failed to create file:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		log.Println("Failed to encode JSON:", err)
		return
	}

	log.Printf("Data saved to %s", filename)
}
