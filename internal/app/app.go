package app

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/config"
	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/logger"
	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/middlewares"
	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/server/route"
	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/server/service"
)

// Application struct
type application struct {
	Server  *http.Server     // the server that processes requests for funds transfer
	Service service.Servicer // service for processing request
	Sigint  chan os.Signal   // channel for given signal for graceful shutdown
}

func New() (application, error) {

	// Create new connect to logger
	log, err := logger.New()
	if err != nil {
		return application{}, err
	}

	// Reading config file
	cfg, err := config.LoadConfig("../")
	if err != nil {
		log.Log.Error("read config file is fail: ", err)
		return application{}, err
	}

	// Initialization service
	service, err := service.New(log, cfg)
	if err != nil {
		return application{}, err
	}

	// Init middlewares
	middlewares := middlewares.New(cfg.AdminSecretKey, cfg.UserSecretKey, log)

	// Init new router
	route := route.New(service, middlewares)

	// Init server
	server := &http.Server{
		Addr:    net.JoinHostPort(cfg.Host, cfg.Port),
		Handler: route,
	}

	// Creating channel for graceful shutdown
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	return application{
		Server: server,
		Sigint: sigint,
	}, nil

}
