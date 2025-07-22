package main

import (
	"fmt"
	"log"
	"net"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"

	"id-generator/internal/config"
	grpcHandler "id-generator/internal/handler/grpc"
	"id-generator/internal/handler/http"
	"id-generator/internal/infra/snowflake"
	"id-generator/internal/usecase"
	pb "id-generator/snowflake-pb"
)

func main() {
	// NOTE: a machine can support a maximum of 4096 new IDs per millisecond.
	conf := config.Init()
	gen, err := snowflake.NewGenerator(int64(*conf.DatacenterID), int64(*conf.MachineID))

	if err != nil {
		log.Fatal("Failed to create Generator:", err)
	}

	uc := usecase.NewIDUsecase(gen)

	go func() {
		app := fiber.New()
		httpHandler := http.NewHandler(uc)
		httpHandler.RegisterRoutes(app)
		log.Fatal(app.Listen(*conf.Listen.Port))
	}()

	lis, err := net.Listen("tcp", *conf.Listen.Grpc)

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterIDServiceServer(s, grpcHandler.NewIDHandler(uc))

	fmt.Println("gRPC server running at ", *conf.Listen.Grpc)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
