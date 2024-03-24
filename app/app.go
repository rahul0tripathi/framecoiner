package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rahul0tripathi/framecoiner/config"
	"github.com/rahul0tripathi/framecoiner/controller"
	"github.com/rahul0tripathi/framecoiner/integrations"
	"github.com/rahul0tripathi/framecoiner/pkg/log"
	"github.com/rahul0tripathi/framecoiner/pkg/redis"
	"github.com/rahul0tripathi/framecoiner/pkg/server"
	"github.com/rahul0tripathi/framecoiner/repo"
	"github.com/rahul0tripathi/framecoiner/services"
	"go.uber.org/zap"
)

func Run() error {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := log.NewZapLogger(false)
	if err != nil {
		return fmt.Errorf("failed to create logger, %w", err)
	}

	cfg, err := config.NewConfigFromEnv()
	if err != nil {
		return fmt.Errorf("failed to get config, %w", err)
	}

	httpserver := server.New(cfg.HostPort, logger)

	storage, _ := redis.NewRedisDB(redis.RedisConfig{
		Addr:     cfg.RedisAddr,
		UserName: cfg.RedisUserName,
		Password: cfg.RedisPassword,
	})

	tradesRepo := repo.NewTradesRepo(storage)
	swapper, err := integrations.NewZeroXSwapper(integrations.ZeroXConfig{
		ApiKey:  cfg.ZeroXApiKey,
		ChainID: cfg.ChainID,
	})
	if err != nil {
		return err
	}

	chainBackend, err := ethclient.Dial(cfg.RpcURL)
	if err != nil {
		return err
	}

	manager := integrations.NewKeyManager(storage)
	processor, err := services.NewTradeProcessor(manager, tradesRepo, swapper, chainBackend, logger, cfg.ChainID)
	if err != nil {
		return err
	}

	accountsSvc := services.NewAccountService(manager, tradesRepo, processor, chainBackend)
	metadataSvc := services.NewTokenMetadataService(chainBackend, integrations.ZeroXConfig{
		ApiKey:  cfg.ZeroXApiKey,
		ChainID: cfg.ChainID,
	})

	processor.Run(ctx, 3)

	controller.SetupRouter(accountsSvc, metadataSvc, httpserver.Router())

	httpserver.Start()

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		logger.Info("M::context canceled")
	case s := <-interrupt:
		logger.Info("M::signal -> " + s.String())
	case err = <-httpserver.Notify():
		return fmt.Errorf("M::notify ->, %w", err)
	}

	err = httpserver.Shutdown()
	if err != nil {
		logger.Error("APP::shutdown, %s", zap.Error(err))
	}

	return nil

}
