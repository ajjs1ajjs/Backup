package datamover

import (
	"bytes"
	"context"
	"io"
)

// RPC Protocol Definitions

type ReadDiskArgs struct {
	SourceURI string
	Offset    int64
	Size      int64
}

type ReadDiskReply struct {
	Data   []byte
	EOF    bool
	Err    string
}

type WriteChunkArgs struct {
	ChunkID string
	Data    []byte
}

type WriteChunkReply struct {
	Success bool
	Err     string
}

type SystemInfoArgs struct{}

type SystemInfoReply struct {
	Info *SystemInfo
	Err  string
}

// Service defines the RPC service export
type Service struct {
	ctx context.Context
	dm  DataMover
}

func NewService(ctx context.Context, dm DataMover) *Service {
	return &Service{ctx: ctx, dm: dm}
}

func (s *Service) GetSystemInfo(args *SystemInfoArgs, reply *SystemInfoReply) error {
	info, err := s.dm.GetSystemInfo(s.ctx)
	if err != nil {
		reply.Err = err.Error()
		return nil
	}
	reply.Info = info
	return nil
}

func (s *Service) WriteChunk(args *WriteChunkArgs, reply *WriteChunkReply) error {
	// Simple bytes.Reader wrapper for the provided data
	data := bytes.NewReader(args.Data)
	err := s.dm.WriteChunk(s.ctx, args.ChunkID, data)
	if err != nil {
		reply.Err = err.Error()
		reply.Success = false
		return nil
	}
	reply.Success = true
	return nil
}

func (s *Service) ReadDisk(args *ReadDiskArgs, reply *ReadDiskReply) error {
	reader, err := s.dm.ReadDisk(s.ctx, args.SourceURI, args.Offset, args.Size)
	if err != nil {
		reply.Err = err.Error()
		return nil
	}
	defer reader.Close()

	data := make([]byte, args.Size)
	n, err := io.ReadFull(reader, data)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		reply.Err = err.Error()
		return nil
	}

	reply.Data = data[:n]
	if err == io.EOF || int64(n) < args.Size {
		reply.EOF = true
	}
	return nil
}
