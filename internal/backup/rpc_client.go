package backup

import (
	"context"
	"fmt"
	"net/rpc"
	"novabackup/internal/datamover"
)

// RemoteDataMoverClient handles RPC communication with a nova-datamover agent
type RemoteDataMoverClient struct {
	client *rpc.Client
}

func NewRemoteDataMoverClient(address string) (*RemoteDataMoverClient, error) {
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to datamover: %w", err)
	}
	return &RemoteDataMoverClient{client: client}, nil
}

func (c *RemoteDataMoverClient) ReadDisk(ctx context.Context, sourceURI string, offset int64, size int64) (string, []byte, bool, error) {
	args := &datamover.ReadDiskArgs{
		SourceURI: sourceURI,
		Offset:    offset,
		Size:      size,
	}
	var reply datamover.ReadDiskReply
	
	err := c.client.Call("Service.ReadDisk", args, &reply)
	if err != nil {
		return "", nil, false, err
	}
	
	if reply.Err != "" {
		return "", nil, false, fmt.Errorf("%s", reply.Err)
	}
	
	return reply.Hash, reply.Data, reply.EOF, nil
}

func (c *RemoteDataMoverClient) Close() error {
	return c.client.Close()
}

func (c *RemoteDataMoverClient) ResetSession() error {
	var reply bool
	arg := "reset"
	return c.client.Call("Service.ResetSession", &arg, &reply)
}

func (c *RemoteDataMoverClient) Ping() error {
	var reply bool
	arg := "ping"
	return c.client.Call("Service.Ping", &arg, &reply)
}
