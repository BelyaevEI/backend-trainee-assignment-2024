package models

// Application constants
const (
	ConfigName string = "app"
	ConfigType string = "env"
)

// Структура общения
type BannerBody struct {
	BannerID  uint32        `json:"banner_id"`
	TagID     uint32        `json:"tag_id"`
	FeatureID uint32        `json:"feature_id"`
	Content   BannerContent `json:"content"`
	Active    bool          `json:"is_active"`
}

// Структура контента
type BannerContent struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Url   string `json:"url"`
}

// Структура ответа
type Response struct {
	Err      error `json:"error"`
	BannerID int   `json:"banner_id"`
}

// Структура запроса
type Query struct {
	FeatureID int
	TagID     int
	Limit     int
	Offset    int
	Last      bool
}

type ResponseBody struct {
	BannerID  uint32        `json:"banner_id"`
	TagID     []uint32      `json:"tag_id"`
	FeatureID uint32        `json:"feature_id"`
	Content   BannerContent `json:"content"`
	Active    bool          `json:"is_active"`
}

// Структура для просмотра истории баннера
type BannerHistory struct {
	BannerID uint32 `json:"banner_id"`
	Version  int    `json:"version"`
	Title    string `json:"title"`
	Text     string `json:"text"`
	Url      string `json:"url"`
}
