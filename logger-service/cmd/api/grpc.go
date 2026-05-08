package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log-service/data"
	"log-service/logs"

	"google.golang.org/grpc"
)

type LogServer struct {
    logs.UnimplementedLogServiceServer
    Models data.Models
}

// WriteLog handles incoming log entries via gRPC
func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
    // Extract log entry from request
    input := req.GetLogEntry()

    // Create log entry for database
    logEntry := data.LogEntry{
        Name:      input.Name,
        Data:      input.Data,
        CreatedAt: time.Now(),
    }

    // Insert log entry into database
    err := l.Models.LogEntry.Insert(logEntry)
    if err != nil {
        log.Printf("Error inserting log: %v", err)
        return &logs.LogResponse{
            Result:  "failed to insert log",
            Success: false,
        }, nil
    }

    // Return success response
    return &logs.LogResponse{
        Result:  "log inserted successfully",
        Success: true,
    }, nil
}

// HealthCheck returns the health status of the service
func (l *LogServer) HealthCheck(ctx context.Context, req *logs.HealthCheckRequest) (*logs.HealthCheckResponse, error) {
    return &logs.HealthCheckResponse{
        Status:  "healthy",
        Version: "1.0.0",
    }, nil
}

func (app *Config) grpcListen() {
    // Create network listener
    lis, err := net.Listen("tcp", fmt.Sprintf(":%s", gRpcPort))
    if err != nil {
        log.Fatalf("Failed to listen for gRPC on port %s: %v", gRpcPort, err)
    }
    defer lis.Close()

    // Create gRPC server with options
    serverOpts := []grpc.ServerOption{
        grpc.MaxRecvMsgSize(1024 * 1024 * 4), // 4MB max message size
        grpc.MaxSendMsgSize(1024 * 1024 * 4), // 4MB max message size
    }
    
    s := grpc.NewServer(serverOpts...)

    // Register LogService server
    logs.RegisterLogServiceServer(s, &LogServer{
        Models: app.Models,
    })

    // Log server startup
    log.Printf("gRPC Server started on port %s", gRpcPort)
    log.Printf("Listening for gRPC connections on tcp://%s", lis.Addr().String())

    // Serve with graceful shutdown
    go func() {
        if err := s.Serve(lis); err != nil {
            log.Fatalf("Failed to serve gRPC: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down gRPC server...")
    s.GracefulStop()
    log.Println("gRPC server stopped")
}