package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "os"
    "time"

    "google.golang.org/grpc"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    consulapi "github.com/hashicorp/consul/api"
    pb "products-service/proto/gen"
)

const serviceName = "products-service"
const servicePort = 50052

type Product struct {
    gorm.Model
    Name  string
    Price float64
}

type server struct {
    pb.UnimplementedProductServiceServer
    db *gorm.DB
}

func (s *server) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductResponse, error) {
    product := Product{Name: req.Name, Price: req.Price}
    if result := s.db.Create(&product); result.Error != nil {
        return nil, result.Error
    }
    return &pb.ProductResponse{Product: &pb.Product{Id: fmt.Sprint(product.ID), Name: product.Name, Price: product.Price}}, nil
}

func (s *server) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.ProductResponse, error) {
    var product Product
    if result := s.db.First(&product, req.Id); result.Error != nil {
        return nil, result.Error
    }
    return &pb.ProductResponse{Product: &pb.Product{Id: fmt.Sprint(product.ID), Name: product.Name, Price: product.Price}}, nil
}

func main() {
    time.Sleep(10 * time.Second)

    dsn := "host=products-db user=user password=password dbname=products_db port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    db.AutoMigrate(&Product{})

    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", servicePort))
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    
    s := grpc.NewServer()
    pb.RegisterProductServiceServer(s, &server{db: db})

    // Register with Consul
    hostname, err := os.Hostname()
    if err != nil {
        log.Fatalf("Failed to get hostname: %v", err)
    }
    
    if err := registerServiceWithConsul(hostname); err != nil {
        log.Printf("Warning: Failed to register with Consul: %v", err)
    } else {
        log.Printf("Successfully registered with Consul as %s", hostname)
    }

    log.Printf("%s gRPC server listening at %v", serviceName, lis.Addr())
    if err := s.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}

func registerServiceWithConsul(hostname string) error {
    config := consulapi.DefaultConfig()
    if addr := os.Getenv("CONSUL_HTTP_ADDR"); addr != "" {
        config.Address = addr
    }
    
    log.Printf("Connecting to Consul at %s", config.Address)

    consul, err := consulapi.NewClient(config)
    if err != nil {
        return fmt.Errorf("failed to create consul client: %w", err)
    }

    registration := &consulapi.AgentServiceRegistration{
        ID:      fmt.Sprintf("%s-%s", serviceName, hostname),
        Name:    serviceName,
        Port:    servicePort,
        Address: hostname,
        Check: &consulapi.AgentServiceCheck{
            TCP:                            fmt.Sprintf("%s:%d", hostname, servicePort),
            Interval:                       "10s",
            Timeout:                        "5s",
            DeregisterCriticalServiceAfter: "30s",
        },
    }
    
    log.Printf("Registering service: %s at %s:%d", serviceName, hostname, servicePort)

    if err := consul.Agent().ServiceRegister(registration); err != nil {
        return fmt.Errorf("failed to register service: %w", err)
    }

    log.Printf("Service registered successfully")
    return nil
}
