package main

import (
	"context"
	"log"
	"time"

	"log-service/data"
	"log-service/logs"
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