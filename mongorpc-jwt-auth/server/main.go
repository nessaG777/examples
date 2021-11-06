package main

import (
	"example/interceptor"
	"fmt"
	"net"
	"os"

	"github.com/mongorpc/mongorpc"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

type MongoRPCConfig struct {
	mongoURI string
	port     int
}

func main() {
	config := &MongoRPCConfig{}

	app := &cli.App{
		Name:  "mongorpc",
		Usage: "mongorpc is a gRPC server that can be used to call mongodb directly",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "mongodb",
				Value:       "mongodb://localhost:27017",
				Usage:       "the mongodb uri",
				Destination: &config.mongoURI,
			},
			&cli.IntFlag{
				Name:        "port",
				Value:       27051,
				Usage:       "the port on which the server will listen",
				Destination: &config.port,
			},
		},
		Action: func(c *cli.Context) error {

			port := fmt.Sprintf(":%d", config.port)

			// connect to mongodb
			database, err := mongo.Connect(c.Context, options.Client().ApplyURI(config.mongoURI))
			if err != nil {
				return err
			}

			mongorpc := &mongorpc.MongoRPC{
				DB: database,
			}

			i := interceptor.Interceptor{
				JWTSecret: "secret",
			}

			srv := mongorpc.NewServer(
				grpc.UnaryInterceptor(i.UnaryInterceptor),
				grpc.StreamInterceptor(i.StreamInterceptor),
			)

			// listen on the port
			listener, err := net.Listen("tcp", port)
			if err != nil {
				return err
			}

			logrus.Printf("mongorpc server is listening at %v", listener.Addr())

			// start the server
			err = srv.Serve(listener)
			return err
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}
