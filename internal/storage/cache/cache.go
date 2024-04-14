package cache

import (
	"strconv"

	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/models"
	"github.com/go-redis/redis"
)

// Implementation check
var _ Cacher = cache{}

type Cacher interface {
	GetBanner(FThash uint64) (models.BannerContent, error)
	SetBanner2Cache(hashKey uint64, banner models.BannerContent)
	DeleteBanner(bannerID int)
}

type cache struct {
	rdb *redis.Client
}

func New(addr, password string) Cacher {

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	return &cache{rdb: rdb}
}

func (c cache) GetBanner(FThash uint64) (models.BannerContent, error) {

	var banner models.BannerContent

	// Получение структуры из хэша Redis
	val, err := c.rdb.HGetAll(strconv.Itoa(int(FThash))).Result()
	if err != nil {
		return models.BannerContent{}, err
	}

	banner.Title = val["Title"]
	banner.Text = val["Text"]
	banner.Url = val["Url"]

	return banner, nil
}
func (c cache) SetBanner2Cache(hashKey uint64, banner models.BannerContent) {
	_ = c.rdb.HSet(strconv.Itoa(int(hashKey)), "key", banner).Err()
}

func (c cache) DeleteBanner(bannerID int) {
}
