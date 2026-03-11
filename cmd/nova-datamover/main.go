package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	"novabackup/internal/datamover"
	"go.uber.org/zap"
)

func main() {
	port := flag.Int("port", 50051, "The server port")
	flag.Parse()

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	logger.Info("Starting Nova Datamover",
		zap.Int("port", *port),
		zap.String("version", "0.1.0"))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	logger.Info("Datamover is listening", zap.String("address", lis.Addr().String()))

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-stop
		logger.Info("Shutting down Datamover...")
		lis.Close()
		cancel()
	}()

	// Initialize DataMover implementation
	dm := datamover.NewDiskReader()
	dedupe := datamover.NewDeduplicator()
	service := datamover.NewService(ctx, dm, dedupe)
	
	rpc.Register(service)

	logger.Info("Datamover RPC service registered", zap.String("address", lis.Addr().String()))

	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					logger.Error("failed to accept connection", zap.Error(err))
					continue
				}
			}
			go rpc.ServeConn(conn)
		}
	}()

	<-ctx.Done()
	logger.Info("Datamover exited")
}
