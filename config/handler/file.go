package handler

import (
	"fmt"
	"github.com/sparrow-community/protos/config"
	"go-micro.dev/v4/config/source"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"os"
	"path"
	"path/filepath"
	"time"
)

// FileService local config file watch
type FileService struct {
	root   string
	memory *Memory
}

func NewFileService(r string) (*FileService, error) {
	memory := NewMemory()
	getPath, err := getPaths(r)
	if err != nil {
		return nil, err
	}
	if err := memory.Watch(getPath...); err != nil {
		return nil, err
	}
	return &FileService{
		root:   r,
		memory: memory,
	}, nil
}

func (f *FileService) Read(ctx context.Context, request *proto.ReadRequest, response *proto.ReadResponse) error {
	p := filepath.Clean(path.Join(f.root, request.Path))
	set, err := f.memory.Get(p)
	if err != nil {
		if err = f.writeFile(p, []byte("")); err != nil {
			return err
		}
	}
	if set == nil {
		set = &source.ChangeSet{}
	}
	response.ChangeSet = &proto.ChangeSet{
		Data:      set.Data,
		Checksum:  set.Checksum,
		Format:    set.Format,
		Source:    set.Source,
		Timestamp: time.Now().Unix(),
	}
	return nil
}

func (f *FileService) Write(_ context.Context, request *proto.WriteRequest, response *wrapperspb.BoolValue) error {
	dest := filepath.Clean(path.Join(f.root, request.Path))
	if err := f.writeFile(dest, request.ChangeSet.Data); err != nil {
		response.Value = false
		return err
	}
	response.Value = true
	return nil
}

func (f *FileService) Watch(_ context.Context, request *proto.WatchRequest, stream proto.Source_WatchStream) error {
	set, err := f.memory.Get(filepath.Clean(path.Join(f.root, request.Path)))
	if err != nil {
		return status.Errorf(codes.NotFound, "cannot read %s", err)
	}
	rsp := &proto.WatchResponse{
		ChangeSet: &proto.ChangeSet{
			Data:      set.Data,
			Checksum:  set.Checksum,
			Format:    set.Format,
			Source:    set.Source,
			Timestamp: time.Now().Unix(),
		},
	}
	if err := stream.Send(rsp); err != nil {
		return status.Errorf(codes.Internal, "watch send response error %s", err)
	}
	return nil
}

func (f *FileService) writeFile(dest string, data []byte) error {
	//dest := filepath.Clean(path.Join(f.root, request.Path))
	dir := filepath.Dir(dest)
	exists, err := exists(dir)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "%s", err)
	}
	if !exists {
		if err := os.MkdirAll(dest, os.ModePerm); err != nil {
			return status.Errorf(codes.InvalidArgument, "MkdirAll error %s %s", dest, err)
		}
	}
	destTmp := fmt.Sprintf("%s.tmp", dest)
	if err := os.WriteFile(destTmp, data, 0666); err != nil {
		return status.Errorf(codes.InvalidArgument, "write file error %s %s", dest, err)
	}
	if err := os.Rename(destTmp, dest); err != nil {
		return status.Errorf(codes.InvalidArgument, "rename %s to %s error %s", destTmp, dest, err)
	}
	if err := f.memory.Watch(dest); err != nil {
		return status.Errorf(codes.InvalidArgument, "watch %s error %s", dest, err)
	}
	return nil
}

// getPaths gets the paths
func getPaths(root string) ([]string, error) {
	if err := os.MkdirAll(root, os.ModePerm); err != nil {
		return nil, err
	}
	var fs []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		fs = append(fs, filepath.Clean(path))
		return nil
	})

	if err != nil {
		return nil, err
	}
	return fs, err
}

func exists(p string) (bool, error) {
	if _, err := os.Stat(p); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return true, nil
	}
	return true, nil
}
