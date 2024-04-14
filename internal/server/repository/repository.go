package repository

import (
	"context"
	"hash/fnv"
	"net/url"
	"strconv"

	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/config"
	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/models"
	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/storage/cache"
	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/storage/database"
)

// Implementation check
var _ Repositorer = (*Repository)(nil)

type Repositorer interface {
	CreateBanner(ctx context.Context, bannerBody models.BannerBody) (int, error)
	UpdateBanner(ctx context.Context, bannerBody models.BannerBody, bannerID int) (bool, error)
	GetQueryParam(querys url.Values) models.Query
	GetBanners(ctx context.Context, queryParam models.Query) ([]models.ResponseBody, error)
	CheckQuery(queryParam models.Query) bool
	GetBanner(ctx context.Context, featureID, tagID int) (models.BannerContent, error)
	GetBannerFromCache(featureID, tagID int) (models.BannerContent, error)
	DeleteBanner(ctx context.Context, bannerID int) error
	GetHistoryBanner(ctx context.Context, bannerID int) ([]models.BannerHistory, error)
	UpdateVersion(ctx context.Context, bannerVersion models.BannerHistory) error
}

// Repository layer
type Repository struct {
	db    database.DBaser
	cache cache.Cacher
}

// Create new repository for service
func New(cfg config.Config) (Repositorer, error) {

	// Connect to postgreSQL database
	postgre, err := database.Connect(cfg.DSN)
	if err != nil {
		return nil, err
	}

	// Connect to redis database
	cache := cache.New(cfg.RedisAddr, cfg.RedisPassword)

	return Repository{
			db:    postgre,
			cache: cache,
		},
		nil
}

func (repo Repository) CreateBanner(ctx context.Context, bannerBody models.BannerBody) (int, error) {
	return repo.db.CreateBanner(ctx, bannerBody)
}

func (repo Repository) UpdateBanner(ctx context.Context, bannerBody models.BannerBody, bannerID int) (bool, error) {
	return repo.db.UpdateBanner(ctx, bannerBody, bannerID)
}

func (repo Repository) GetQueryParam(querys url.Values) models.Query {

	var d models.Query

	if val, ok := querys["feature_id"]; ok {
		d.TagID, _ = strconv.Atoi(val[0])
	}

	if val, ok := querys["tag_id"]; ok {
		d.TagID, _ = strconv.Atoi(val[0])
	}

	if val, ok := querys["limit"]; ok {
		d.Limit, _ = strconv.Atoi(val[0])
	}

	if val, ok := querys["offset"]; ok {
		d.Offset, _ = strconv.Atoi(val[0])
	}

	if val, ok := querys["use_last_revision"]; ok {
		d.Last, _ = strconv.ParseBool(val[0])
	}

	return d

}

func (repo Repository) GetBanners(ctx context.Context, queryParam models.Query) ([]models.ResponseBody, error) {

	// Получим все баннеры из БД по параметрам
	banners, err := repo.db.GetBanners(ctx, queryParam)
	if err != nil {
		return []models.ResponseBody{}, err
	}

	// Отфильтруем по переданным limit и offset

	return banners, nil
}

func (repo Repository) CheckQuery(queryParam models.Query) bool {
	if queryParam.FeatureID != 0 && queryParam.TagID != 0 {
		return true
	}
	return false
}

func (repo Repository) GetBanner(ctx context.Context, featureID, tagID int) (models.BannerContent, error) {

	banner, err := repo.db.GetBanner(ctx, featureID, tagID)
	if err != nil {
		return models.BannerContent{}, err
	}

	// Запишем баннер в кэш
	repo.setBanner2Cache(repo.hashTwoFields(featureID, tagID), banner)

	return banner, nil
}

func (repo Repository) GetBannerFromCache(featureID, tagID int) (models.BannerContent, error) {

	// Преобразуем фичу и тэг в хэш, чтобы записать в кэш
	return repo.cache.GetBanner(repo.hashTwoFields(featureID, tagID))
}

func (repo Repository) hashTwoFields(featureID, tagID int) uint64 {
	h := fnv.New64a()
	h.Write([]byte(strconv.Itoa(featureID)))
	h.Write([]byte(strconv.Itoa(tagID)))
	return h.Sum64()
}

func (repo Repository) setBanner2Cache(hashKey uint64, banner models.BannerContent) {
	repo.cache.SetBanner2Cache(hashKey, banner)
}

func (repo Repository) DeleteBanner(ctx context.Context, bannerID int) error {

	//Удаляем из БД
	err := repo.db.DeleteBanner(ctx, bannerID)
	if err != nil {
		return err
	}

	// Удаляем из кэща
	repo.cache.DeleteBanner(bannerID)

	return nil
}

func (repo Repository) GetHistoryBanner(ctx context.Context, bannerID int) ([]models.BannerHistory, error) {
	return repo.db.GetHistoryBanner(ctx, bannerID)
}

func (repo Repository) UpdateVersion(ctx context.Context, bannerVersion models.BannerHistory) error {
	return repo.db.UpdateVersion(ctx, bannerVersion)
}
