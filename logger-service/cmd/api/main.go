package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"log-service/data"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	webPort  = "8080"
	mongoURL = "mongodb://admin:password@mongo:27017/logs?authSource=admin"
	rpcPort   = "5001"
    gRpcPort  = "50001"
)

type Config struct {
	Models data.Models
	Mongo  *mongo.Client
}

func main() {
	client, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}

	app := Config{
		Models: data.New(client),
		Mongo:  client,
	}

	log.Println("Connected to MongoDB!")

	// Register the RPC Server
	rpcServer := &RPCServer{
		Client: app.Mongo,
	}

	err = rpc.Register(rpcServer)
	if err != nil {
		log.Panic(err)
	}

	go app.rpcListen()
	
	err = app.serve()
	if err != nil {
		log.Panic(err)
	}
}

func (app *Config) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", webPort),
		Handler:      app.routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting server on port %s", webPort)
	return srv.ListenAndServe()
}

func connectToMongo() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(mongoURL)

	c, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = c.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (app *Config) rpcListen() error {
	log.Println("Starting RPC server on port ", rpcPort)
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))
	if err != nil {
		return err
	}
	defer listen.Close()

	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(rpcConn)
	}
}