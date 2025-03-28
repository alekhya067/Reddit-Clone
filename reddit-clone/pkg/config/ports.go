// pkg/config/ports.go
package config

import "fmt"

const (
    // Default ports for different services
    DefaultEnginePort  = 50051  // Main gRPC server port
    DefaultMetricsPort = 50052  // Metrics server port
    DefaultClientPort  = 50053  // Client metrics port
)

// ServiceConfig holds configuration for all services
type ServiceConfig struct {
    // Server addresses
    EngineHost  string
    EnginePort  int
    MetricsPort int
    ClientPort  int

    // Additional configuration if needed
    MaxConnections    int
    ConnectionTimeout int
}

// NewDefaultConfig creates a ServiceConfig with default values
func NewDefaultConfig() *ServiceConfig {
    return &ServiceConfig{
        EngineHost:        "localhost",
        EnginePort:        DefaultEnginePort,
        MetricsPort:       DefaultMetricsPort,
        ClientPort:        DefaultClientPort,
        MaxConnections:    1000,
        ConnectionTimeout: 30,
    }
}

// GetEngineAddress returns the complete engine server address
func (c *ServiceConfig) GetEngineAddress() string {
    return fmt.Sprintf("%s:%d", c.EngineHost, c.EnginePort)
}

// GetMetricsAddress returns the complete metrics server address
func (c *ServiceConfig) GetMetricsAddress() string {
    return fmt.Sprintf("%s:%d", c.EngineHost, c.MetricsPort)
}

// GetClientMetricsAddress returns the client metrics server address
func (c *ServiceConfig) GetClientMetricsAddress() string {
    return fmt.Sprintf("%s:%d", c.EngineHost, c.ClientPort)
}

// Validate checks if the configuration is valid
func (c *ServiceConfig) Validate() error {
    if c.EnginePort == c.MetricsPort || c.EnginePort == c.ClientPort {
        return fmt.Errorf("ports must be different")
    }
    if c.MaxConnections < 1 {
        return fmt.Errorf("max connections must be positive")
    }
    return nil
}