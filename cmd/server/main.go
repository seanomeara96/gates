package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/seanomeara96/gates/config"
	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repos/cache"
	"github.com/seanomeara96/gates/repos/sqlite"
	"github.com/seanomeara96/gates/router"
)

func server() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// ROUTING LOGIC
	// middleware executed in reverse order; i = 0 executes last
	router, err := router.DefaultRouter(cfg)
	if err != nil {
		return err
	}
	defer router.Close()

	srv := http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router.Mux(),
	}

	errChan := make(chan error)
	go func() {
		log.Println("Listening on http://localhost:" + cfg.Port)
		if err := srv.ListenAndServe(); err != nil {
			errChan <- fmt.Errorf("server startup: failed to listen and serve on env.config.Port %s: %w", cfg.Port, err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		close(errChan)
		return fmt.Errorf("server errored: %w", err)
	case <-quit:
		close(quit)
		log.Println("shutting down server gracefully")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("server forced to shut down %w", err)
		}
		log.Println("server shut down gracefully")
		return nil
	}
}

func main() {
	if err := server(); err != nil {
		log.Fatal(err)
	}
}

/* service funcs start */

func TotalValue(products *cache.CachedProductRepo, cart models.Cart) (float32, error) {
	var value float32 = 0.0
	for _, item := range cart.Items {
		for _, component := range item.Components {
			productPrice, err := products.GetProductPrice(component.Product.Id)
			if err != nil {
				return 0, err
			}
			value += (productPrice * float32(component.Qty))
		}
	}
	return value, nil
}
func RemoveItem(cartRepo *sqlite.CartRepo, cartID, itemID string) error {
	// TODO add transactions
	if err := cartRepo.RemoveCartItem(cartID, itemID); err != nil {
		return fmt.Errorf("failed to remove item. %w", err)
	}
	if err := cartRepo.RemoveCartItemComponents(itemID); err != nil {
		return fmt.Errorf("failed to remove item components. %w", err)
	}
	return nil
}
