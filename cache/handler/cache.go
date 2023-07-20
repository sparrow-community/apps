package handler

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sparrow-community/protos/cache"
	"time"
)

type Cache struct {
	Client *redis.Client
}

func (c Cache) Get(ctx context.Context, in *cache.GetRequest, out *cache.GetResponse) error {
	v, err := c.Client.Get(ctx, in.Key).Result()
	if err != nil && err == redis.Nil {
		return nil
	} else if err != nil {
		return err
	}
	dur, err := c.Client.TTL(ctx, in.Key).Result()
	if err != nil {
		return err
	}
	out.Key = in.Key
	out.Value = v
	out.Ttl = time.Now().Add(dur).Unix()
	return nil
}

func (c Cache) Set(ctx context.Context, in *cache.SetRequest, out *cache.SetResponse) error {
	ret, err := c.Client.Set(ctx, in.Key, in.Value, time.Duration(in.Ttl)*time.Second).Result()
	if err != nil {
		return err
	}
	out.Status = ret
	return nil
}

func (c Cache) Delete(ctx context.Context, in *cache.DeleteRequest, out *cache.DeleteResponse) error {
	ret, err := c.Client.Del(ctx, in.Key).Result()
	if err != nil {
		return err
	}
	out.Status = fmt.Sprintf("%d", ret)
	return nil
}

func (c Cache) Increment(ctx context.Context, in *cache.IncrementRequest, out *cache.IncrementResponse) error {
	ret, err := c.Client.IncrBy(ctx, in.Key, in.Value).Result()
	if err != nil {
		return err
	}
	out.Key = in.Key
	out.Value = ret
	return nil
}

func (c Cache) Decrement(ctx context.Context, in *cache.DecrementRequest, out *cache.DecrementResponse) error {
	ret, err := c.Client.DecrBy(ctx, in.Key, in.Value).Result()
	if err != nil {
		return err
	}
	out.Key = in.Key
	out.Value = ret
	return nil
}

func (c Cache) ListKeys(ctx context.Context, _ *cache.ListKeysRequest, out *cache.ListKeysResponse) error {
	ret, err := c.Client.Keys(ctx, "*").Result()
	if err != nil {
		return err
	}
	out.Keys = ret
	return nil
}

func (c Cache) HGet(ctx context.Context, in *cache.HGetRequest, out *cache.HGetResponse) error {
	ret, err := c.Client.HGet(ctx, in.Key, in.Field).Result()
	if err != nil && err == redis.Nil {
		return nil
	} else if err != nil {
		return err
	}
	out.Key = in.Key
	out.Field = in.Field
	out.Value = ret
	return nil
}

func (c Cache) HSet(ctx context.Context, in *cache.HSetRequest, out *cache.HSetResponse) error {
	ret, err := c.Client.HSet(ctx, in.Key, in.Field, in.Value).Result()
	if err != nil {
		return err
	}
	out.Status = fmt.Sprintf("%d", ret)
	return nil
}

func (c Cache) HSetMap(ctx context.Context, in *cache.HSetMapRequest, out *cache.HSetMapResponse) error {
	ret, err := c.Client.HSet(ctx, in.Key, in.Value).Result()
	if err != nil {
		return err
	}
	out.Status = fmt.Sprintf("%d", ret)
	return nil
}

func (c Cache) HGetAll(ctx context.Context, in *cache.HGetAllRequest, out *cache.HGetAllResponse) error {
	ret, err := c.Client.HGetAll(ctx, in.Key).Result()
	if err != nil && err == redis.Nil {
		return nil
	} else if err != nil {
		return err
	}
	out.Value = ret
	return nil
}
