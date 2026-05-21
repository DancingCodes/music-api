# music-api 方案

## 依赖

| 库 | 作用 |
|---|---|
| `godotenv` | 加载 `.env` 文件 |
| `gin` | HTTP 路由 + 参数解析 + JSON 返回 |
| `gorm` + `mysql driver` | 操作 MySQL，不用写 SQL |
| `cos-go-sdk-v5` | 上传音频到腾讯云 COS |

## 目录结构

```
music-api/
  go.mod
  .env
  main.go
  handler.go          # HTTP handler
  service.go          # 业务逻辑
  model.go            # Music 结构体
  netease.go          # 网易云 DTO
  db.go               # GORM 连接
  log.go              # 自定义 slog handler
  response.go         # 统一响应
  httpclient.go       # GetJSON / 上传 COS
```

## 路由

```
GET  /net/search    # 搜索网易云
POST /music/save    # 保存歌曲
GET  /music/list    # 歌曲列表
```

## 配置

环境变量，`os.Getenv` 读取：

| 变量 | 说明 |
|---|---|
| `dbDSN` | MySQL 连接串 |
| `neteaseCookie` | 网易云登录 Cookie |
| `cosSecretID` | 腾讯云 SecretId |
| `cosSecretKey` | 腾讯云 SecretKey |
| `cosBucketURL` | COS 存储桶地址 |
| `cosPathPrefix` | COS 上传目录前缀，如 `music/` |
| `cosCDNURL` | CDN 加速域名，可选。不设则用 COS 原始地址 |

## 数据表

```go
type Music struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    Name        string    `gorm:"type:varchar(255);index;not null" json:"name"`
    Url         string    `gorm:"type:varchar(1024)" json:"url"`
    PicUrl      string    `gorm:"type:varchar(500)" json:"pic_url"`
    Artists     string    `gorm:"type:varchar(255)" json:"artists"`
    DurationMs  int       `gorm:"column:duration_ms;type:int" json:"duration_ms"`
    Lyric       string    `gorm:"type:text" json:"lyric"`
    CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}
```

## 网易云 API

```go
// 搜索
url := "https://music.163.com/api/search/get/web?s=" + keyword + "&type=1&offset=" + offset + "&limit=" + limit

// 详情
url := fmt.Sprintf("https://music.163.com/api/v3/song/detail?id=%d&c=[{id:%d}]", id, id)

// 歌词
url := fmt.Sprintf("https://music.163.com/api/song/lyric?id=%d&lv=-1&tv=-1", id)

// 播放地址（需要 Cookie）
url := fmt.Sprintf("https://music.163.com/api/song/enhance/player/url/v1?ids=[%d]&encodeType=aac&level=jymaster", id)
headers := map[string]string{"Cookie": NETEASE_COOKIE}
```

/music/save 流程：三个接口并发请求，拿到 COS 地址后写库。库中已有则直接返回。

## 响应格式

```go
// 成功
{ "code": 200, "msg": "success", "data": {...} }

// 失败
{ "code": 500, "msg": "错误信息", "data": null }
```
