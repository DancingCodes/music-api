package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"sync"
)

func SearchNetease(name string, pageNo, pageSize int) (map[string]any, error) {
	offset := (pageNo - 1) * pageSize
	apiUrl := fmt.Sprintf(
		"https://music.163.com/api/search/get/web?s=%s&type=1&offset=%d&limit=%d",
		url.QueryEscape(name), offset, pageSize,
	)

	raw, err := GetJSON[map[string]any](apiUrl, nil)
	if err != nil {
		return nil, err
	}

	return raw["result"].(map[string]any), nil
}

func SaveMusicLogic(songID int) (*Music, error) {
	var existing Music
	if err := DB.Where("id = ?", songID).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("歌曲 %s 已存在", existing.Name)
	}

	neteaseCookie := os.Getenv("neteaseCookie")

	var (
		wg                sync.WaitGroup
		detailRes         DetailResponse
		lyricRes          LyricResponse
		urlRes            URLResponse
		detailErr, urlErr error
	)

	wg.Add(3)

	go func() {
		defer wg.Done()
		apiUrl := fmt.Sprintf("https://music.163.com/api/v3/song/detail?id=%d&c=[{id:%d}]", songID, songID)
		detailRes, detailErr = GetJSON[DetailResponse](apiUrl, nil)
	}()

	go func() {
		defer wg.Done()
		apiUrl := fmt.Sprintf("https://music.163.com/api/song/lyric?id=%d&lv=-1&tv=-1", songID)
		lyricRes, _ = GetJSON[LyricResponse](apiUrl, nil)
	}()

	go func() {
		defer wg.Done()
		apiUrl := fmt.Sprintf("https://music.163.com/api/song/enhance/player/url/v1?ids=[%d]&encodeType=aac&level=standard", songID)
		urlRes, urlErr = GetJSON[URLResponse](apiUrl, map[string]string{
			"Cookie": neteaseCookie,
		})
	}()

	wg.Wait()

	if detailErr != nil || len(detailRes.Songs) == 0 {
		return nil, fmt.Errorf("获取详情失败")
	}
	if urlErr != nil || len(urlRes.Data) == 0 || urlRes.Data[0].URL == "" {
		return nil, fmt.Errorf("获取播放链接失败或版权受限")
	}

	song := detailRes.Songs[0]
	audioURL := urlRes.Data[0].URL
	fileType := urlRes.Data[0].Type

	cosPathPrefix := os.Getenv("cosPathPrefix")
	cosKey := fmt.Sprintf("%s%d.%s", cosPathPrefix, songID, fileType)
	cosURL, err := UploadToCOS(audioURL, cosKey)
	if err != nil {
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	var artists []string
	for _, ar := range song.Ar {
		artists = append(artists, ar.Name)
	}

	newMusic := Music{
		ID:         uint(song.ID),
		Name:       song.Name,
		Url:        cosURL,
		PicUrl:     song.Al.PicURL,
		Artists:    strings.Join(artists, ","),
		DurationMs: song.Dt,
		Lyric:      lyricRes.Lrc.Lyric,
	}

	if err := DB.Create(&newMusic).Error; err != nil {
		return nil, fmt.Errorf("数据库入库失败: %w", err)
	}

	slog.Info("歌曲已保存", "歌曲ID", songID, "歌名", newMusic.Name)
	return &newMusic, nil
}

func GetMusicListLogic(pageNo, pageSize int, name string) ([]Music, int64, error) {
	var musicList []Music
	var total int64

	query := DB.Model(&Music{})
	if name != "" {
		keyword := "%" + name + "%"
		query = query.Where("name LIKE ? OR artists LIKE ?", keyword, keyword)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNo - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id desc").Find(&musicList).Error

	return musicList, total, err
}

func DeleteMusicLogic(id int) error {
	var music Music
	if err := DB.Where("id = ?", id).First(&music).Error; err != nil {
		return fmt.Errorf("歌曲不存在")
	}

	if err := DB.Delete(&music).Error; err != nil {
		return fmt.Errorf("删除失败: %w", err)
	}

	if cosClientObj != nil && music.Url != "" {
		u, err := url.Parse(music.Url)
		if err == nil {
			objectKey := strings.TrimPrefix(u.Path, "/")
			if _, err := cosClientObj.Object.Delete(context.Background(), objectKey); err != nil {
				slog.Error("COS 文件删除失败", "objectKey", objectKey, "错误", err)
			}
		}
	}

	return nil
}
