package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Twitter Snowflake constants
const (
	// Twitter epoch: Nov 04, 2010, 01:42:54 UTC (in milliseconds)
	TwitterEpoch = 1288834974657

	// Bit allocation
	TimestampBits    = 41
	DatacenterIdBits = 5
	MachineIdBits    = 5
	SequenceBits     = 12

	// Max values
	MaxDatacenterId = (1 << DatacenterIdBits) - 1 // 31
	MaxMachineId    = (1 << MachineIdBits) - 1    // 31
	MaxSequence     = (1 << SequenceBits) - 1     // 4095

	// Bit shifts
	MachineIdShift    = SequenceBits                                    // 12
	DatacenterIdShift = SequenceBits + MachineIdBits                    // 17
	TimestampShift    = SequenceBits + MachineIdBits + DatacenterIdBits // 22
)

// SnowflakeGenerator generates unique IDs using the Snowflake algorithm
type SnowflakeGenerator struct {
	mu           sync.Mutex
	datacenterId int64
	machineId    int64
	sequence     int64
	lastTime     int64
}

// NewSnowflakeGenerator creates a new Snowflake generator
func NewSnowflakeGenerator(datacenterId, machineId int64) (*SnowflakeGenerator, error) {
	if datacenterId < 0 || datacenterId > MaxDatacenterId {
		return nil, fmt.Errorf("datacenter ID must be between 0 and %d", MaxDatacenterId)
	}
	if machineId < 0 || machineId > MaxMachineId {
		return nil, fmt.Errorf("machine ID must be between 0 and %d", MaxMachineId)
	}

	return &SnowflakeGenerator{
		datacenterId: datacenterId,
		machineId:    machineId,
		sequence:     0,
		lastTime:     -1,
	}, nil
}

// NextID generates the next unique ID
func (sg *SnowflakeGenerator) NextID() (int64, error) {
	sg.mu.Lock()
	defer sg.mu.Unlock()

	timestamp := sg.timeGen()

	// Clock moved backwards
	if timestamp < sg.lastTime {
		return 0, fmt.Errorf("clock moved backwards. Refusing to generate ID for %d milliseconds", sg.lastTime-timestamp)
	}

	// Same millisecond as last ID generation
	if timestamp == sg.lastTime {
		sg.sequence = (sg.sequence + 1) & MaxSequence
		if sg.sequence == 0 {
			// Sequence exhausted, wait for next millisecond
			timestamp = sg.tilNextMillis(sg.lastTime)
		}
	} else {
		sg.sequence = 0
	}

	sg.lastTime = timestamp

	// Generate ID
	id := ((timestamp - TwitterEpoch) << TimestampShift) |
		(sg.datacenterId << DatacenterIdShift) |
		(sg.machineId << MachineIdShift) |
		sg.sequence

	return id, nil
}

// timeGen returns current timestamp in milliseconds
func (sg *SnowflakeGenerator) timeGen() int64 {
	return time.Now().UnixNano() / 1e6
}

// tilNextMillis waits until next millisecond
func (sg *SnowflakeGenerator) tilNextMillis(lastTimestamp int64) int64 {
	timestamp := sg.timeGen()
	for timestamp <= lastTimestamp {
		timestamp = sg.timeGen()
	}
	return timestamp
}

// ParseID parses a snowflake ID into its components
func (sg *SnowflakeGenerator) ParseID(id int64) map[string]interface{} {
	timestamp := (id >> TimestampShift) + TwitterEpoch
	datacenterId := (id >> DatacenterIdShift) & ((1 << DatacenterIdBits) - 1)
	machineId := (id >> MachineIdShift) & ((1 << MachineIdBits) - 1)
	sequence := id & ((1 << SequenceBits) - 1)

	return map[string]interface{}{
		"id":            id,
		"timestamp":     timestamp,
		"datetime":      time.Unix(timestamp/1000, (timestamp%1000)*1e6).UTC().Format(time.RFC3339),
		"datacenter_id": datacenterId,
		"machine_id":    machineId,
		"sequence":      sequence,
	}
}

// gRPC Protocol Buffer definitions (normally in a separate .proto file)
// For this example, we'll define the structures directly

type GenerateIdRequest struct{}

type GenerateIdResponse struct {
	Id int64 `json:"id"`
}

type ParseIdRequest struct {
	Id int64 `json:"id"`
}

type ParseIdResponse struct {
	Id           int64  `json:"id"`
	Timestamp    int64  `json:"timestamp"`
	Datetime     string `json:"datetime"`
	DatacenterId int64  `json:"datacenter_id"`
	MachineId    int64  `json:"machine_id"`
	Sequence     int64  `json:"sequence"`
}

// gRPC Server implementation
type IdGeneratorServer struct {
	generator *SnowflakeGenerator
}

func (s *IdGeneratorServer) GenerateId(ctx context.Context, req *GenerateIdRequest) (*GenerateIdResponse, error) {
	id, err := s.generator.NextID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate ID: %v", err)
	}
	return &GenerateIdResponse{Id: id}, nil
}

func (s *IdGeneratorServer) ParseId(ctx context.Context, req *ParseIdRequest) (*ParseIdResponse, error) {
	parsed := s.generator.ParseID(req.Id)

	return &ParseIdResponse{
		Id:           parsed["id"].(int64),
		Timestamp:    parsed["timestamp"].(int64),
		Datetime:     parsed["datetime"].(string),
		DatacenterId: parsed["datacenter_id"].(int64),
		MachineId:    parsed["machine_id"].(int64),
		Sequence:     parsed["sequence"].(int64),
	}, nil
}

// Service holds both Fiber and gRPC components
type Service struct {
	generator  *SnowflakeGenerator
	fiberApp   *fiber.App
	grpcServer *grpc.Server
}

// NewService creates a new service instance
func NewService(datacenterId, machineId int64) (*Service, error) {
	generator, err := NewSnowflakeGenerator(datacenterId, machineId)
	if err != nil {
		return nil, err
	}

	// Setup Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Snowflake ID Generator",
	})

	// Middleware
	app.Use(cors.New())
	app.Use(logger.New())

	service := &Service{
		generator: generator,
		fiberApp:  app,
	}

	// Setup REST routes
	service.setupRoutes()

	// Setup gRPC server
	service.grpcServer = grpc.NewServer()
	// Register gRPC service (normally done with generated code)
	// For this example, we'll simulate it

	return service, nil
}

// setupRoutes configures Fiber routes
func (s *Service) setupRoutes() {
	api := s.fiberApp.Group("/api/v1")

	// Generate single ID
	api.Get("/id", s.generateId)

	// Generate multiple IDs
	api.Get("/ids/:count", s.generateIds)

	// Parse ID
	api.Get("/parse/:id", s.parseId)

	// Health check
	api.Get("/health", s.health)

	// Stats
	api.Get("/stats", s.stats)
}

// REST handlers
func (s *Service) generateId(c *fiber.Ctx) error {
	id, err := s.generator.NextID()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"id": id,
	})
}

func (s *Service) generateIds(c *fiber.Ctx) error {
	count := c.Params("count")
	if count == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "count parameter is required",
		})
	}

	var countInt int
	if _, err := fmt.Sscanf(count, "%d", &countInt); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid count parameter",
		})
	}

	if countInt <= 0 || countInt > 10000 {
		return c.Status(400).JSON(fiber.Map{
			"error": "count must be between 1 and 10000",
		})
	}

	ids := make([]int64, countInt)
	for i := 0; i < countInt; i++ {
		id, err := s.generator.NextID()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("failed to generate ID at index %d: %v", i, err),
			})
		}
		ids[i] = id
	}

	return c.JSON(fiber.Map{
		"ids":   ids,
		"count": len(ids),
	})
}

func (s *Service) parseId(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	var id int64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid id parameter",
		})
	}

	parsed := s.generator.ParseID(id)
	return c.JSON(parsed)
}

func (s *Service) health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}

func (s *Service) stats(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"datacenter_id": s.generator.datacenterId,
		"machine_id":    s.generator.machineId,
		"max_sequence":  MaxSequence,
		"twitter_epoch": TwitterEpoch,
	})
}

// StartHTTP starts the Fiber HTTP server
func (s *Service) StartHTTP(port string) error {
	log.Printf("Starting HTTP server on port %s", port)
	return s.fiberApp.Listen(":" + port)
}

// StartGRPC starts the gRPC server
func (s *Service) StartGRPC(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %v", port, err)
	}

	// Register the service
	grpcService := &IdGeneratorServer{generator: s.generator}
	_ = grpcService // Register with gRPC server when proto definitions are available

	log.Printf("Starting gRPC server on port %s", port)
	return s.grpcServer.Serve(lis)
}

// Benchmark function to test performance
func (s *Service) BenchmarkGeneration(count int) (time.Duration, error) {
	start := time.Now()

	for i := 0; i < count; i++ {
		_, err := s.generator.NextID()
		if err != nil {
			return 0, err
		}
	}

	return time.Since(start), nil
}

func main() {
	// Configuration - in production, these would come from environment variables
	datacenterId := int64(1)
	machineId := int64(1)
	httpPort := "8080"
	grpcPort := "9090"

	// Create service
	service, err := NewService(datacenterId, machineId)
	if err != nil {
		log.Fatal("Failed to create service:", err)
	}

	// Performance test
	log.Println("Running performance test...")
	duration, err := service.BenchmarkGeneration(10000)
	if err != nil {
		log.Printf("Benchmark failed: %v", err)
	} else {
		rate := float64(10000) / duration.Seconds()
		log.Printf("Generated 10,000 IDs in %v (%.2f IDs/second)", duration, rate)
	}

	// Start servers concurrently
	go func() {
		if err := service.StartGRPC(grpcPort); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// Start HTTP server (blocking)
	if err := service.StartHTTP(httpPort); err != nil {
		log.Fatal("HTTP server error:", err)
	}
}
