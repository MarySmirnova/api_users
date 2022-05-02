package internal

import (
	"net/http"

	"github.com/MarySmirnova/api_users/internal/api"
	"github.com/MarySmirnova/api_users/internal/config"
	"github.com/MarySmirnova/api_users/internal/database"

	log "github.com/sirupsen/logrus"
)

type Application struct {
	cfg config.Application
	db  *database.DB
}

func NewApplication(cfg config.Application) (*Application, error) {
	app := &Application{
		cfg: cfg,
	}

	if err := app.initDatabase(); err != nil {
		return nil, err
	}

	return app, nil
}

func (a *Application) initDatabase() error {
	db := database.New()
	err := db.NewUser(&database.User{
		Username: a.cfg.AdminUsername,
		Password: a.cfg.AdminPass,
		Admin:    true,
	})
	if err != nil {
		return err
	}

	a.db = db
	return nil
}

func (a *Application) StartServer() {
	srv := api.New(a.cfg.API, a.db)
	s := srv.GetHTTPServer()

	log.WithField("listen", s.Addr).Info("start server")

	err := s.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		log.WithError(err).Error("the channel raised an error")
		return
	}
}
