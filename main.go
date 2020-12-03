package router_runner

import (
	"context"
	"github.com/labstack/echo/v4"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	HOST = may("HOST", "0.0.0.0")
	PORT = may("PORT", "8080")
)

func may(envName, defaultValue string) string {
	envValue := os.Getenv(envName)
	if envValue == "" {
		return defaultValue
	}
	return envValue
}

func RunWithRoutes(initCtx context.Context, routers ...func(context.Context) func(context.Context, *echo.Echo) error) (context.Context, error) {
	ctx, cancel := context.WithCancel(initCtx)

	e := echo.New()
	for _, routerFactory := range routers {
		if err := routerFactory(ctx)(ctx, e); err != nil {
			cancel()
			log.Printf("Error in setting routes: %v $([1]T)", err)
			return nil, err
		}
	}

	go func() {
		defer cancel()
		sigCh := make(chan os.Signal)
		signal.Notify(sigCh,
			syscall.SIGINT,
			syscall.SIGHUP,
			syscall.SIGTERM,
			syscall.SIGUSR2)
		sig := <-sigCh
		signal.Reset(sig)
		log.Printf("Got signal %q, terminating...", sig)

		ctx, _ = context.WithTimeout(ctx, 5*time.Second)
		if err := e.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown cause: %v (%[1]T)", err)
		}
	}()

	go func() {
		defer cancel()
		log.Printf("Server listenig at %s:%s", HOST, PORT)
		if err := e.Start(HOST + ":" + PORT); err != nil {
			log.Printf("HTTP server termination cause: %v (%[1]T)", err)
		}
	}()

	return ctx, nil
}
