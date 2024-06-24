package service

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/config"
	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/logger"
	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/models"
	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/server/repository"
	"github.com/go-chi/chi"
)

// Implementation check
var _ Servicer = (*Service)(nil)

type Servicer interface {
	GetUserBanner(writer http.ResponseWriter, request *http.Request)
	CreateBanner(writer http.ResponseWriter, request *http.Request)
	UpdateBanner(writer http.ResponseWriter, request *http.Request)
	GetBanners(writer http.ResponseWriter, request *http.Request)
	DeleteBanner(writer http.ResponseWriter, request *http.Request)
	GetHistoryBanner(writer http.ResponseWriter, request *http.Request)
	UpdateVersion(writer http.ResponseWriter, request *http.Request)
}

type Service struct {
	log        *logger.Logger
	repository repository.Repositorer
	Route      *chi.Mux
}

func New(log *logger.Logger, cfg config.Config) (Servicer, error) {

	// Create a new repository
	repository, err := repository.New(cfg)
	if err != nil {
		return &Service{}, err
	}

	return &Service{
		log:        log,
		repository: repository,
	}, nil
}

func (s *Service) GetUserBanner(writer http.ResponseWriter, request *http.Request) {

	var response models.Response

	ctx := request.Context()

	writer.Header().Set("Content-Type", "application/json")

	// Получим параметры запроса
	queryParam := s.repository.GetQueryParam(request.URL.Query())

	// Необходимо проверить, что переданные данные в запросе не пустые
	if ok := s.repository.CheckQuery(queryParam); !ok {
		s.log.Log.Error("error read body request")
		writer.WriteHeader(http.StatusBadRequest)
		response.Err = errors.New("error read body request")
		json.NewEncoder(writer).Encode(response)
		return
	}

	// Получим баннер из БД
	if queryParam.Last {
		banner, err := s.repository.GetBanner(ctx, queryParam.FeatureID, queryParam.TagID)
		if err != nil {
			if banner == (models.BannerContent{}) {
				s.log.Log.Error("banner not found")
				writer.WriteHeader(http.StatusNotFound)
				response.Err = errors.New("banner not found")
				json.NewEncoder(writer).Encode(response)
				return
			}
			s.log.Log.Error("geting banner is failed: ", err)
			writer.WriteHeader(http.StatusInternalServerError)
			response.Err = err
			json.NewEncoder(writer).Encode(response)
			return
		}

		// Если все ОК
		writer.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(writer).Encode(banner); err != nil {
			s.log.Log.Error("searilizing banners is failed: ", err)
		}
		return
	}

	// Получим баннер из кэша
	banner, err := s.repository.GetBannerFromCache(queryParam.FeatureID, queryParam.TagID)
	if err != nil {
		s.log.Log.Error("geting banner is failed: ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		response.Err = err
		json.NewEncoder(writer).Encode(response)
		return
	}

	writer.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(writer).Encode(banner); err != nil {
		s.log.Log.Error("searilizing banners is failed: ", err)
	}

}

// Эндпойнт должен создавать баннер, которого нет в системе с уникальной парой фича + тэг
func (s *Service) CreateBanner(writer http.ResponseWriter, request *http.Request) {

	var (
		bannerBody models.BannerBody
		response   models.Response
		bannerID   int
	)

	ctx := request.Context()

	writer.Header().Set("Content-Type", "application/json")

	// Читаем тело запроса
	body, err := io.ReadAll(request.Body)
	if err != nil {
		s.log.Log.Error("error read body request: ", err)
		writer.WriteHeader(http.StatusBadRequest)
		response.Err = err
		json.NewEncoder(writer).Encode(response)
		return
	}

	// Десериализуем JSON
	if err = json.Unmarshal(body, &bannerBody); err != nil {
		s.log.Log.Error("unmarshal json is failed: ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		response.Err = err
		json.NewEncoder(writer).Encode(response)
		return
	}

	// Создаем баннер
	if bannerID, err = s.repository.CreateBanner(ctx, bannerBody); err != nil {
		s.log.Log.Error("creating banner is failed: ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		response.Err = err
		json.NewEncoder(writer).Encode(response)
		return
	}

	// Если все ОК, отвечаем ID баннера
	writer.WriteHeader(http.StatusCreated)
	response.BannerID = bannerID
	json.NewEncoder(writer).Encode(response)

}

// Обновление содержимого баннера
func (s *Service) UpdateBanner(writer http.ResponseWriter, request *http.Request) {
	var (
		bannerBody models.BannerBody
		response   models.Response
		ok         bool
	)

	ctx := request.Context()

	writer.Header().Set("Content-Type", "application/json")

	// Получим ID из url запроса
	// Т.е. мы используем chi можем использовать chi.URLParam, но в тестах это не работает
	path := request.URL.Path
	parts := strings.Split(path, "/")
	bannerID, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		s.log.Log.Error("reading banner id from request is failed: ", err)
		writer.WriteHeader(http.StatusBadRequest)
		response.Err = err
		json.NewEncoder(writer).Encode(response)
		return
	}

	// Читаем тело запроса
	body, err := io.ReadAll(request.Body)
	if err != nil {
		s.log.Log.Error("error read body request: ", err)
		writer.WriteHeader(http.StatusBadRequest)
		response.Err = err
		json.NewEncoder(writer).Encode(response)
		return
	}

	// Десериализуем JSON
	if err = json.Unmarshal(body, &bannerBody); err != nil {
		s.log.Log.Error("unmarshal json is failed: ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		response.Err = err
		json.NewEncoder(writer).Encode(response)
		return
	}

	// Обновляем баннер
	if ok, err = s.repository.UpdateBanner(ctx, bannerBody, bannerID); err != nil {
		s.log.Log.Error("updating banner is failed: ", err)
		writer.WriteHeader(http.StatusBadRequest)
		response.Err = err
		json.NewEncoder(writer).Encode(response)
		return
	}

	// Баннер не найден
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	// Если все ОК
	writer.WriteHeader(http.StatusOK)

}

// Получение всех баннеров по переданным параметрам
func (s *Service) GetBanners(writer http.ResponseWriter, request *http.Request) {

	var response models.Response

	ctx := request.Context()

	writer.Header().Set("Content-Type", "application/json")

	// Получим параметры запроса
	queryParam := s.repository.GetQueryParam(request.URL.Query())

	// Получим баннеры по условиям запроса
	banners, err := s.repository.GetBanners(ctx, queryParam)
	if err != nil {
		s.log.Log.Error("updating banner is failed: ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		response.Err = err
		json.NewEncoder(writer).Encode(response)
		return
	}

	// Если все ОК
	writer.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(writer).Encode(banners); err != nil {
		s.log.Log.Error("searilizing banners is failed: ", err)
	}

}

// Удаление баннера
func (s *Service) DeleteBanner(writer http.ResponseWriter, request *http.Request) {

	var response models.Response

	ctx := request.Context()

	writer.Header().Set("Content-Type", "application/json")

	path := request.URL.Path
	parts := strings.Split(path, "/")
	bannerID, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		s.log.Log.Error("reading banner id from request is failed: ", err)
		writer.WriteHeader(http.StatusBadRequest)
		response.Err = err
		json.NewEncoder(writer).Encode(response)
		return
	}

	// удалим баннер
	if err := s.repository.DeleteBanner(ctx, bannerID); err != nil {
		s.log.Log.Error("deleting banner is failed: ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		response.Err = err
		json.NewEncoder(writer).Encode(response)
		return
	}

	// Если все ОК
	writer.WriteHeader(http.StatusNoContent)
}

// Просмотр всей истории баннера
func (s *Service) GetHistoryBanner(writer http.ResponseWriter, request *http.Request) {

	var (
		response      models.Response
		bannerHistory []models.BannerHistory
	)

	ctx := request.Context()

	writer.Header().Set("Content-Type", "application/json")

	path := request.URL.Path
	parts := strings.Split(path, "/")
	bannerID, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		s.log.Log.Error("reading banner id from request is failed: ", err)
		writer.WriteHeader(http.StatusBadRequest)
		response.Err = err
		json.NewEncoder(writer).Encode(response)
		return
	}

	// вернем всю историю баннера
	if bannerHistory, err = s.repository.GetHistoryBanner(ctx, bannerID); err != nil {
		s.log.Log.Error("getting history banner is failed: ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		response.Err = err
		json.NewEncoder(writer).Encode(response)
		return
	}

	// Если все ОК
	writer.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(writer).Encode(bannerHistory); err != nil {
		s.log.Log.Error("searilizing banners is failed: ", err)
	}

}

// Обновление версии баннера
func (s *Service) UpdateVersion(writer http.ResponseWriter, request *http.Request) {

	var (
		response      models.Response
		bannerVersion models.BannerHistory
	)

	ctx := request.Context()

	writer.Header().Set("Content-Type", "application/json")

	// Читаем тело запроса
	body, err := io.ReadAll(request.Body)
	if err != nil {
		s.log.Log.Error("error read body request: ", err)
		writer.WriteHeader(http.StatusBadRequest)
		response.Err = err
		json.NewEncoder(writer).Encode(response)
		return
	}

	// Десериализуем JSON
	if err = json.Unmarshal(body, &bannerVersion); err != nil {
		s.log.Log.Error("unmarshal json is failed: ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		response.Err = err
		json.NewEncoder(writer).Encode(response)
		return
	}

	// Обновляем версию баннера
	if err = s.repository.UpdateVersion(ctx, bannerVersion); err != nil {
		s.log.Log.Error("updating version is failed: ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		response.Err = err
		json.NewEncoder(writer).Encode(response)
		return
	}

}
