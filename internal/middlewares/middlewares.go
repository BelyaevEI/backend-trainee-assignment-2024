package middlewares

import (
	"fmt"
	"net/http"

	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/logger"
	"github.com/golang-jwt/jwt"
)

type Middlewares struct {
	adminSecretKey string // Admin secret key
	userSecretKey  string // User secret key
	log            *logger.Logger
}

func New(adminSecretKey, userSecretKey string, log *logger.Logger) *Middlewares {
	return &Middlewares{
		adminSecretKey: adminSecretKey,
		log:            log,
		userSecretKey:  userSecretKey,
	}
}

func (middlewares *Middlewares) AdminAuthorization(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		ctx := request.Context()

		token := request.Header["Token"][0]
		if len(token) == 0 {
			middlewares.log.Log.Error("token is empty")
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		// Validation admin token
		ok, err := middlewares.validation(token, middlewares.adminSecretKey)
		if err != nil {
			middlewares.log.Log.Errorf("token validation is failed: ", err)
			writer.WriteHeader(http.StatusForbidden)
			return
		}

		if !ok {
			middlewares.log.Log.Info("token validation is failed")
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(writer, request.WithContext(ctx))
	})
}

func (middlewares *Middlewares) UserAuthorization(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		ctx := request.Context()

		token := request.Header["Token"][0]
		if len(token) == 0 {
			middlewares.log.Log.Error("token is empty")
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		// Пробуем сначала аутентифицировать по UserToken
		ok, err := middlewares.validation(token, middlewares.userSecretKey)
		if err != nil {
			middlewares.log.Log.Errorf("token validation is failed: ", err)
			writer.WriteHeader(http.StatusForbidden)
			return
		}

		// Если переданный токен не user проверяем на admin, у них тоже есть доступ
		if !ok {
			ok, err = middlewares.validation(token, middlewares.adminSecretKey)
			if err != nil {
				middlewares.log.Log.Errorf("token validation is failed: ", err)
				writer.WriteHeader(http.StatusForbidden)
				return
			}
			if !ok {
				middlewares.log.Log.Info("token validation is failed")
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		h.ServeHTTP(writer, request.WithContext(ctx))

	})
}

func (middlewares *Middlewares) validation(token, secretkey string) (bool, error) {
	tok, err := jwt.Parse(token,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretkey), nil
		})

	if !tok.Valid {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
