package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"zero-music/config"
	"zero-music/models"
	"zero-music/services"

	"github.com/gin-gonic/gin"
)

// SearchHandler 搜索处理器
type SearchHandler struct {
	scanner services.Scanner
}

// NewSearchHandler 创建搜索处理器
func NewSearchHandler(scanner services.Scanner) *SearchHandler {
	return &SearchHandler{scanner: scanner}
}

// SearchResult 搜索结果
type SearchResult struct {
	Songs   []*models.Song `json:"songs"`
	Total   int            `json:"total"`
	Artists []string       `json:"artists,omitempty"`
	Albums  []string       `json:"albums,omitempty"`
}

// Search 综合搜索
func (h *SearchHandler) Search(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "搜索关键词不能为空",
		})
		return
	}

	searchType := c.DefaultQuery("type", "all") // all, song, artist, album
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit <= 0 || limit > config.MaxSearchLimit {
		limit = config.DefaultSearchLimit
	}
	if offset < 0 {
		offset = 0
	}

	songs := h.scanner.GetSongs()
	queryLower := strings.ToLower(query)

	var matchedSongs []*models.Song
	artistSet := make(map[string]bool)
	albumSet := make(map[string]bool)

	for _, song := range songs {
		matched := false

		switch searchType {
		case "song":
			if containsIgnoreCase(song.Title, queryLower) {
				matched = true
			}
		case "artist":
			if containsIgnoreCase(song.Artist, queryLower) {
				matched = true
			}
		case "album":
			if containsIgnoreCase(song.Album, queryLower) {
				matched = true
			}
		default: // all
			if containsIgnoreCase(song.Title, queryLower) ||
				containsIgnoreCase(song.Artist, queryLower) ||
				containsIgnoreCase(song.Album, queryLower) {
				matched = true
			}
		}

		if matched {
			matchedSongs = append(matchedSongs, song)
			if song.Artist != "" {
				artistSet[song.Artist] = true
			}
			if song.Album != "" {
				albumSet[song.Album] = true
			}
		}
	}

	// 按相关性排序（标题匹配优先）
	sort.Slice(matchedSongs, func(i, j int) bool {
		titleMatchI := containsIgnoreCase(matchedSongs[i].Title, queryLower)
		titleMatchJ := containsIgnoreCase(matchedSongs[j].Title, queryLower)
		if titleMatchI && !titleMatchJ {
			return true
		}
		if !titleMatchI && titleMatchJ {
			return false
		}
		return matchedSongs[i].Title < matchedSongs[j].Title
	})

	total := len(matchedSongs)

	// 分页
	if offset >= total {
		matchedSongs = []*models.Song{}
	} else {
		end := offset + limit
		if end > total {
			end = total
		}
		matchedSongs = matchedSongs[offset:end]
	}

	// 收集唯一的艺术家和专辑
	var artists, albums []string
	for artist := range artistSet {
		artists = append(artists, artist)
	}
	for album := range albumSet {
		albums = append(albums, album)
	}
	sort.Strings(artists)
	sort.Strings(albums)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": SearchResult{
			Songs:   matchedSongs,
			Total:   total,
			Artists: artists,
			Albums:  albums,
		},
	})
}

// GetArtists 获取所有艺术家列表
func (h *SearchHandler) GetArtists(c *gin.Context) {
	songs := h.scanner.GetSongs()

	artistMap := make(map[string]int) // 艺术家 -> 歌曲数量
	for _, song := range songs {
		if song.Artist != "" {
			artistMap[song.Artist]++
		}
	}

	type ArtistInfo struct {
		Name      string `json:"name"`
		SongCount int    `json:"song_count"`
	}

	var artists []ArtistInfo
	for name, count := range artistMap {
		artists = append(artists, ArtistInfo{Name: name, SongCount: count})
	}

	// 按歌曲数量排序
	sort.Slice(artists, func(i, j int) bool {
		if artists[i].SongCount != artists[j].SongCount {
			return artists[i].SongCount > artists[j].SongCount
		}
		return artists[i].Name < artists[j].Name
	})

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"artists": artists,
			"total":   len(artists),
		},
	})
}

// GetArtistSongs 获取指定艺术家的歌曲
func (h *SearchHandler) GetArtistSongs(c *gin.Context) {
	artist := c.Param("name")
	if artist == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "艺术家名称不能为空",
		})
		return
	}

	songs := h.scanner.GetSongs()
	var artistSongs []*models.Song
	albumSet := make(map[string]bool)

	for _, song := range songs {
		if strings.EqualFold(song.Artist, artist) {
			artistSongs = append(artistSongs, song)
			if song.Album != "" {
				albumSet[song.Album] = true
			}
		}
	}

	// 按专辑和标题排序
	sort.Slice(artistSongs, func(i, j int) bool {
		if artistSongs[i].Album != artistSongs[j].Album {
			return artistSongs[i].Album < artistSongs[j].Album
		}
		return artistSongs[i].Title < artistSongs[j].Title
	})

	var albums []string
	for album := range albumSet {
		albums = append(albums, album)
	}
	sort.Strings(albums)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"artist": artist,
			"songs":  artistSongs,
			"albums": albums,
			"total":  len(artistSongs),
		},
	})
}

// GetAlbums 获取所有专辑列表
func (h *SearchHandler) GetAlbums(c *gin.Context) {
	songs := h.scanner.GetSongs()

	type AlbumInfo struct {
		Name      string `json:"name"`
		Artist    string `json:"artist"`
		SongCount int    `json:"song_count"`
		Year      int    `json:"year,omitempty"`
	}

	albumMap := make(map[string]*AlbumInfo)
	for _, song := range songs {
		if song.Album == "" {
			continue
		}
		key := song.Album + "|" + song.Artist
		if info, exists := albumMap[key]; exists {
			info.SongCount++
		} else {
			albumMap[key] = &AlbumInfo{
				Name:      song.Album,
				Artist:    song.Artist,
				SongCount: 1,
				Year:      song.Year,
			}
		}
	}

	var albums []*AlbumInfo
	for _, info := range albumMap {
		albums = append(albums, info)
	}

	// 按专辑名排序
	sort.Slice(albums, func(i, j int) bool {
		return albums[i].Name < albums[j].Name
	})

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"albums": albums,
			"total":  len(albums),
		},
	})
}

// GetAlbumSongs 获取指定专辑的歌曲
func (h *SearchHandler) GetAlbumSongs(c *gin.Context) {
	album := c.Param("name")
	if album == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "专辑名称不能为空",
		})
		return
	}

	songs := h.scanner.GetSongs()
	var albumSongs []*models.Song
	var artist string
	var year int

	for _, song := range songs {
		if strings.EqualFold(song.Album, album) {
			albumSongs = append(albumSongs, song)
			if artist == "" && song.Artist != "" {
				artist = song.Artist
			}
			if year == 0 && song.Year > 0 {
				year = song.Year
			}
		}
	}

	// 按曲目号排序（如果有的话），否则按标题
	sort.Slice(albumSongs, func(i, j int) bool {
		if albumSongs[i].Track != albumSongs[j].Track {
			return albumSongs[i].Track < albumSongs[j].Track
		}
		return albumSongs[i].Title < albumSongs[j].Title
	})

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"album":  album,
			"artist": artist,
			"year":   year,
			"songs":  albumSongs,
			"total":  len(albumSongs),
		},
	})
}

// containsIgnoreCase 忽略大小写检查字符串包含
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), substr)
}
