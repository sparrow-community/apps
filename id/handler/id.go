package handler

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/mattheath/kala/bigflake"
	"github.com/mattheath/kala/snowflake"
	"github.com/oklog/ulid"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
	pb "github.com/sparrow-community/protos/id"
	"github.com/teris-io/shortid"
	"go-micro.dev/v4/errors"
	"go-micro.dev/v4/logger"
	"math/rand"
	"time"
)

type Id struct {
	Snowflake *snowflake.Snowflake
	Bigflake  *bigflake.Bigflake
}

func New() *Id {
	id := rand.Intn(100)

	sf, err := snowflake.New(uint32(id))
	if err != nil {
		panic(err.Error())
	}
	bg, err := bigflake.New(uint64(id))
	if err != nil {
		panic(err.Error())
	}

	return &Id{
		Snowflake: sf,
		Bigflake:  bg,
	}
}

func (id *Id) Generate(ctx context.Context, req *pb.GenerateRequest, rsp *pb.GenerateResponse) error {
	switch req.Type {

	case pb.Types_NANOID: // unsortable
		id, err := gonanoid.New(21) // custom length id!
		if err != nil {
			logger.Errorf("Failed to generate nanoid id: %v", err)
			return errors.InternalServerError("id.generate", "failed to generate nano id")
		}
		rsp.Type = req.Type
		rsp.Id = id
	case pb.Types_ULID: // sortable
		t := time.Now().UTC()
		entropy := rand.New(rand.NewSource(t.UnixNano()))
		id := ulid.MustNew(ulid.Timestamp(t), entropy)
		rsp.Type = req.Type
		rsp.Id = id.String()
	case pb.Types_KSUID: // sortable
		rsp.Type = pb.Types_KSUID
		rsp.Id = ksuid.New().String()
	case pb.Types_XID:
		rsp.Type = req.Type
		rsp.Id = xid.New().String()
	case pb.Types_UUID:
		rsp.Type = req.Type
		rsp.Id = uuid.New().String()
	case pb.Types_SNOWFLAKE:
		id, err := id.Snowflake.Mint()
		if err != nil {
			logger.Errorf("Failed to generate snowflake id: %v", err)
			return errors.InternalServerError("id.generate", "failed to mint snowflake id")
		}
		rsp.Type = req.Type
		rsp.Id = fmt.Sprintf("%v", id)
	case pb.Types_BIGFLAKE:
		id, err := id.Bigflake.Mint()
		if err != nil {
			logger.Errorf("Failed to generate bigflake id: %v", err)
			return errors.InternalServerError("id.generate", "failed to mint bigflake id")
		}
		rsp.Type = req.Type
		rsp.Id = fmt.Sprintf("%v", id)
	case pb.Types_SHORTID:
		id, err := shortid.Generate()
		if err != nil {
			logger.Errorf("Failed to generate shortid id: %v", err)
			return errors.InternalServerError("id.generate", "failed to generate short id")
		}
		rsp.Type = req.Type
		rsp.Id = id
	default:
		return errors.BadRequest("id.generate", "unsupported id type")
	}

	return nil
}

func (id *Id) Types(ctx context.Context, req *pb.TypesRequest, rsp *pb.TypesResponse) error {
	for _, v := range pb.Types_name {
		rsp.Types = append(rsp.Types, v)
	}
	return nil
}
