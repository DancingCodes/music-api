package main

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api")
	{
		api.GET("/net/search", GetNetSearch)
		api.POST("/music/save", SaveMusic)
		api.GET("/music/list", GetMusicList)
		api.DELETE("/music/delete", DeleteMusic)
	}

	return r
}

func GetNetSearch(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		Error(c, "请输入搜索关键词")
		return
	}

	pageNo, _ := strconv.Atoi(c.DefaultQuery("pageNo", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	list, total, err := SearchNetease(name, pageNo, pageSize)
	if err != nil {
		slog.Error("搜索失败", "错误", err)
		Error(c, "搜索失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

func SaveMusic(c *gin.Context) {
	var req struct {
		ID int `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, "参数校验失败，请传入有效的 id")
		return
	}

	music, err := SaveMusicLogic(req.ID)
	if err != nil {
		slog.Error("保存歌曲失败", "歌曲ID", req.ID, "错误", err)
		Error(c, err.Error())
		return
	}

	Success(c, music)
}

func GetMusicList(c *gin.Context) {
	pageNo, _ := strconv.Atoi(c.DefaultQuery("pageNo", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	name := c.Query("name")

	list, total, err := GetMusicListLogic(pageNo, pageSize, name)
	if err != nil {
		slog.Error("获取列表失败", "错误", err)
		Error(c, "获取列表失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

func DeleteMusic(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil || id <= 0 {
		Error(c, "请传入有效的歌曲 id")
		return
	}

	if err := DeleteMusicLogic(id); err != nil {
		slog.Error("删除歌曲失败", "歌曲ID", id, "错误", err)
		Error(c, err.Error())
		return
	}

	Success(c, nil)
}
