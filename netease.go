package main

// DetailResponse 详情响应
type DetailResponse struct {
	Songs []DetailSong `json:"songs"`
}

type DetailSong struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Dt   int    `json:"dt"`
	Al   struct {
		PicURL string `json:"picUrl"`
	} `json:"al"`
	Ar []struct {
		Name string `json:"name"`
	} `json:"ar"`
}

// LyricResponse 歌词响应
type LyricResponse struct {
	Lrc struct {
		Lyric string `json:"lyric"`
	} `json:"lrc"`
}

// URLResponse 播放地址响应
type URLResponse struct {
	Data []URLData `json:"data"`
}

type URLData struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}
