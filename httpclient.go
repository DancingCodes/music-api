package main

import (
	"context"
	"encoding/json"
	"fmt"
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
		return
	}
	u, err := url.Parse(bucketURL)
	if err != nil {
		panic(err)
	}
	cosClientObj = cos.NewClient(&cos.BaseURL{BucketURL: u}, &stdhttp.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  os.Getenv("cosSecretID"),
			SecretKey: os.Getenv("cosSecretKey"),
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

	if resp.StatusCode >= 400 {
		return result, fmt.Errorf("http error: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func UploadToCOS(audioURL string, objectKey string) (string, error) {
	resp, err := httpClientObj.Get(audioURL)
	if err != nil {
		return "", fmt.Errorf("download audio failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != stdhttp.StatusOK {
		return "", fmt.Errorf("audio url returned status: %d", resp.StatusCode)
	}

	_, err = cosClientObj.Object.Put(context.Background(), objectKey, resp.Body, nil)
	if err != nil {
		return "", fmt.Errorf("cos upload failed: %w", err)
	}

	baseURL := os.Getenv("cosBucketURL")
	if cdn := os.Getenv("cosCDNURL"); cdn != "" {
		baseURL = cdn
	}
	return baseURL + "/" + objectKey, nil
}
