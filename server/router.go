package server

import (
	"fmt"
	"net/http"

	"github.com/Scalingo/go-handlers"
	"github.com/Scalingo/sclng-backend-test-v1/config"
	"github.com/sirupsen/logrus"
)

func InitRouter(cfg *config.Config, log logrus.FieldLogger) error {
	log.Info("Initializing routes")
	router := handlers.NewRouter(log)
	router.HandleFunc("/ping", PongHandler)
	router.HandleFunc("/repos", GetRepoHandler)
	router.HandleFunc("/stats", GetStatsHandler)

	log = log.WithField("port", cfg.Port)
	log.Info("Listening on port: ", cfg.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), router)
	if err != nil {
		log.WithError(err).Error("Fail to listen to the given port")
		return err
	}
	return nil
}
