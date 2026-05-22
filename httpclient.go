package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

var httpClientObj = &http.Client{Timeout: 10 * time.Second}

var downloadClient = &http.Client{Timeout: 5 * time.Minute}

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
	cosClientObj = cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})
}

func GetJSON[T any](urlStr string, headers map[string]string) (T, error) {
	var result T

	req, err := http.NewRequest("GET", urlStr, nil)
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
	resp, err := downloadClient.Get(audioURL)
	if err != nil {
		return "", fmt.Errorf("音频下载失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("音频地址返回异常状态码: %d", resp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentLength: resp.ContentLength,
		},
	}
	_, err = cosClientObj.Object.Put(ctx, objectKey, resp.Body, opt)
	if err != nil {
		return "", fmt.Errorf("COS 上传失败: %w", err)
	}

	baseURL := os.Getenv("cosBucketURL")
	if cdn := os.Getenv("cosCDNURL"); cdn != "" {
		baseURL = cdn
	}
	return baseURL + "/" + objectKey, nil
}
