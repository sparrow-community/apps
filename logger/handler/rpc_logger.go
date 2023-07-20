package handler

import (
	"context"
	"fmt"
	"github.com/sparrow-community/protos/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"path/filepath"
	"strings"
)

type logFile struct {
	file *os.File
	stop chan bool
}

type LoggerService struct {
	root     string
	logFiles map[string]logFile
}

func (s *LoggerService) Write(_ context.Context, request *proto.WriteRequest, response *proto.WriteResponse) error {
	f, ok := s.logFiles[request.ServiceName]
	if !ok {
		name := strings.Replace(request.ServiceName, "/", "-", -1)
		dest := filepath.Join(s.root, fmt.Sprintf("%v.log", name))
		dir := filepath.Dir(dest)
		if ex, err := exists(dir); err != nil {
			return status.Errorf(codes.Internal, "%s", err)
		} else if !ex {
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				return status.Errorf(codes.Internal, "MkdirAll error %s %s", dest, err)
			}
		}
		file, err := os.OpenFile(dest, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return status.Errorf(codes.Internal, "Open log file error %s %s", dest, err)
		}
		f = logFile{file, make(chan bool)}
		s.logFiles[request.ServiceName] = f
	}
	n, err := f.file.Write(request.Data)
	if err != nil {
		if f.stop != nil {
			close(f.stop)
		}
		delete(s.logFiles, request.ServiceName)
		return status.Errorf(codes.Internal, "Write log error %s %s", request.ServiceName, err)
	}
	response.N = int32(n)
	return nil
}

func NewLoggerService(root string) *LoggerService {
	return &LoggerService{root: root, logFiles: make(map[string]logFile)}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
