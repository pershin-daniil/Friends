package main

import (
	"Friends/config"
	"Friends/logg"
	"Friends/server"
	"Friends/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	lg := logg.New()
	lg.Info("start server")

	confg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		lg.Error("load config err", "error", err)
	}

	psql, err := storage.New(lg,
		confg.App.Development.Database.Username,
		confg.App.Development.Database.Password,
		confg.App.Development.Database.Address,
		confg.App.Development.Database.NameDatabase,
	)
	if err != nil {
		lg.Error("Failed to connect to database",
			"error", err)
		return
	}

	defer func() {
		if err = psql.Close(); err != nil {
			lg.Error("Failed to close",
				"error", err)
		}
	}()
	psql.MigriteUP()
	httpServer := server.NewServer(lg, confg.App.Development.Server.HTTPPort, psql)

	go func() {

		if err = httpServer.Run(); err != nil {
			lg.Error("Server failed to start", "error", err)
			return
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{})

	go func() {
		<-stop
		err = httpServer.ShutDown()
		if err != nil {
			lg.Error("Failed to shutdown gracefully", "error", err)
		}
		lg.Info("Server gracefully stopped", "error", err)
		close(done)
	}()
	<-done
}
