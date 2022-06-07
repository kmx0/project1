package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kmx0/project1/cmd/gophermart/handlers"
	"github.com/kmx0/project1/cmd/gophermart/storage"
	"github.com/kmx0/project1/internal/config"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetReportCaller(true)
	globalCtx := context.Background()
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	// return
	// ttl := time.Now().Add(time.Second)
	// time.Sleep(time.Second*5)
	// logrus.Info(ttl.After(time.Now()))
	// return
	exitChan := make(chan int)

	go func() {
		for {
			s := <-signalChanel
			switch s {
			// kill -SIGHUP XXXX [XXXX - идентификатор процесса для программы]
			case syscall.SIGINT:
				logrus.Info("Signal interrupt triggered.")
				exitChan <- 0
				// kill -SIGTERM XXXX [XXXX - идентификатор процесса для программы]
			case syscall.SIGTERM:
				logrus.Info("Signal terminte triggered.")
				exitChan <- 0

				// kill -SIGQUIT XXXX [XXXX - идентификатор процесса для программы]
			case syscall.SIGQUIT:
				logrus.Info("Signal quit triggered.")
				exitChan <- 0

			default:
				logrus.Info("Unknown signal.")
				exitChan <- 1
			}
		}
	}()
	cfg := config.LoadConfig()
	logrus.Infof("CFG for SERVER  %+v", cfg)
	storage.PingDB(globalCtx, cfg.DBURI)
	r := handlers.SetupRouter(cfg)
	go http.ListenAndServe(cfg.Address, r)
	exitCode := <-exitChan

	// globalCtx.Done()
	logrus.Warn("Exiting with code ", exitCode)
	os.Exit(exitCode)
}
