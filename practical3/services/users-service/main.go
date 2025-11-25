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
    pb "users-service/proto/gen"
)

const serviceName = "users-service"
const servicePort = 50051

type User struct {
    gorm.Model
    Name  string
    Email string `gorm:"unique"`
}

type server struct {
    pb.UnimplementedUserServiceServer
    db *gorm.DB
}

func (s *server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
    user := User{Name: req.Name, Email: req.Email}
    if result := s.db.Create(&user); result.Error != nil {
        return nil, result.Error
    }
    return &pb.UserResponse{User: &pb.User{Id: fmt.Sprint(user.ID), Name: user.Name, Email: user.Email}}, nil
}

func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
    var user User
    if result := s.db.First(&user, req.Id); result.Error != nil {
        return nil, result.Error
    }
    return &pb.UserResponse{User: &pb.User{Id: fmt.Sprint(user.ID), Name: user.Name, Email: user.Email}}, nil
}

func main() {
    time.Sleep(10 * time.Second)

    dsn := "host=users-db user=user password=password dbname=users_db port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    db.AutoMigrate(&User{})

    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", servicePort))
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    
    s := grpc.NewServer()
    pb.RegisterUserServiceServer(s, &server{db: db})

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
