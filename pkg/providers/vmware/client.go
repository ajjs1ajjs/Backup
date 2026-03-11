// Package vmware provides VMware vSphere integration for NovaBackup
package vmware

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/zap"
)

// Client represents a VMware vSphere client
type Client struct {
	logger     *zap.Logger
	client     *govmomi.Client
	finder     *find.Finder
	datacenter *object.Datacenter
	ctx        context.Context
	cancel     context.CancelFunc
}

// ConnectionConfig holds connection parameters for vCenter/ESXi
type ConnectionConfig struct {
	Host       string // vCenter or ESXi hostname/IP
	Port       int    // Default: 443
	Username   string
	Password   string
	Insecure   bool   // Skip SSL certificate verification
	Datacenter string // Optional: specific datacenter
}

// DefaultConnectionConfig returns default connection settings
func DefaultConnectionConfig() *ConnectionConfig {
	return &ConnectionConfig{
		Port:     443,
		Insecure: true, // Development default
	}
}

// NewClient creates a new VMware client
func NewClient(logger *zap.Logger, config *ConnectionConfig) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		logger: logger.With(zap.String("component", "vmware-client")),
		ctx:    ctx,
		cancel: cancel,
	}

	if err := client.connect(config); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to vSphere: %w", err)
	}

	return client, nil
}

// connect establishes connection to vCenter/ESXi
func (c *Client) connect(config *ConnectionConfig) error {
	u := &url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:%d", config.Host, config.Port),
		User:   url.UserPassword(config.Username, config.Password),
		Path:   "/sdk",
	}

	soapClient := soap.NewClient(u, config.Insecure)
	vimClient, err := vim25.NewClient(c.ctx, soapClient)
	if err != nil {
		return fmt.Errorf("failed to create vim client: %w", err)
	}

	// Create govmomi client
	c.client = &govmomi.Client{
		Client: vimClient,
	}

	// Authenticate using session manager
	userInfo := url.UserPassword(config.Username, config.Password)
	if err := c.client.Login(c.ctx, userInfo); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	c.logger.Info("Successfully connected to vSphere",
		zap.String("host", config.Host),
		zap.String("username", config.Username),
	)

	// Initialize finder for inventory navigation
	c.finder = find.NewFinder(c.client.Client, false)

	// Set datacenter if specified
	if config.Datacenter != "" {
		dc, err := c.finder.Datacenter(c.ctx, config.Datacenter)
		if err != nil {
			return fmt.Errorf("datacenter not found: %w", err)
		}
		c.datacenter = dc
		c.finder.SetDatacenter(dc)
	}

	return nil
}

// Close disconnects from vSphere
func (c *Client) Close() error {
	if c.client != nil {
		if err := c.client.SessionManager.Logout(c.ctx); err != nil {
			c.logger.Warn("Failed to logout from vSphere", zap.Error(err))
		}
	}
	c.cancel()
	return nil
}

// IsConnected checks if client is connected
func (c *Client) IsConnected() bool {
	if c.client == nil {
		return false
	}
	// Try a simple API call to verify connection
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	var serviceInstance mo.ServiceInstance
	err := c.client.RetrieveOne(ctx, types.ManagedObjectReference{
		Type:  "ServiceInstance",
		Value: "ServiceInstance",
	}, nil, &serviceInstance)
	return err == nil
}

// GetDatacenter returns the current datacenter
func (c *Client) GetDatacenter() *object.Datacenter {
	return c.datacenter
}

// SetDatacenter sets the active datacenter
func (c *Client) SetDatacenter(name string) error {
	dc, err := c.finder.Datacenter(c.ctx, name)
	if err != nil {
		return fmt.Errorf("datacenter not found: %w", err)
	}
	c.datacenter = dc
	c.finder.SetDatacenter(dc)
	return nil
}

// GetClient returns the underlying govmomi client
func (c *Client) GetClient() *govmomi.Client {
	return c.client
}

// GetFinder returns the inventory finder
func (c *Client) GetFinder() *find.Finder {
	return c.finder
}

// GetContext returns the client context
func (c *Client) GetContext() context.Context {
	return c.ctx
}
