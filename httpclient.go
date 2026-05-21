package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"net/url"
	"os"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

var httpClientObj = &stdhttp.Client{Timeout: 10 * time.Second}

var cosClientObj *cos.Client

func initCOS() {
	bucketURL := os.Getenv("cosBucketURL")
	if bucketURL == "" {
		slog.Error("cosBucketURL 未设置")
		os.Exit(1)
	}
	secretID := os.Getenv("cosSecretID")
	if secretID == "" {
		slog.Error("cosSecretID 未设置")
		os.Exit(1)
	}
	secretKey := os.Getenv("cosSecretKey")
	if secretKey == "" {
		slog.Error("cosSecretKey 未设置")
		os.Exit(1)
	}

	u, err := url.Parse(bucketURL)
	if err != nil {
		slog.Error("cosBucketURL 解析失败", "错误", err)
		os.Exit(1)
	}
	cosClientObj = cos.NewClient(&cos.BaseURL{BucketURL: u}, &stdhttp.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})
}

func GetJSON[T any](urlStr string, headers map[string]string) (T, error) {
	var result T

	req, err := stdhttp.NewRequest("GET", urlStr, nil)
	if err != nil {
		return result, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := httpClientObj.Do(req)
	if err != nil {
		return result, err
	}
	defer func() { _ = resp.Body.Close() }()

	return result, json.NewDecoder(resp.Body).Decode(&result)
}

func UploadToCOS(audioURL string, objectKey string) (string, error) {
	resp, err := httpClientObj.Get(audioURL)
	if err != nil {
		return "", fmt.Errorf("音频下载失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != stdhttp.StatusOK {
		return "", fmt.Errorf("音频地址返回异常状态码: %d", resp.StatusCode)
	}

	_, err = cosClientObj.Object.Put(context.Background(), objectKey, resp.Body, nil)
	if err != nil {
		return "", fmt.Errorf("COS 上传失败: %w", err)
	}

	baseURL := os.Getenv("cosBucketURL")
	if cdn := os.Getenv("cosCDNURL"); cdn != "" {
		baseURL = cdn
	}
	return baseURL + "/" + objectKey, nil
}
