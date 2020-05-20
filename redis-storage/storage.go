package redis_storage

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/dutchcoders/transfer.sh/server"
	"github.com/go-redis/redis/v7"
)

type RedisStorage struct {
	client *redis.Client
}

const (
	redisTypeSubKey   = "type"
	redisLengthSubKey = "length"
)

func New(addr, password string) server.Storage {
	client := redis.NewClient(&redis.Options{Addr: addr, Password: password})
	return &RedisStorage{
		client: client,
	}
}

func (r *RedisStorage) Get(token string, filename string) (reader io.ReadCloser, contentLength uint64, err error) {
	var val []byte
	if val, err = r.client.Get(formRedisKey(token, filename)).Bytes(); err != nil {
		return
	}
	if contentLength, err = r.Head(token, filename); err != nil {
		return
	}
	reader = ioutil.NopCloser(bytes.NewReader(val))
	return
}

func (r *RedisStorage) Head(token string, filename string) (contentLength uint64, err error) {
	return r.client.Get(fmt.Sprintf("%s:%s", formRedisKey(token, filename), redisLengthSubKey)).Uint64()
}

func (r *RedisStorage) Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) error {
	key := formRedisKey(token, filename)
	r.client.Set(key+":"+redisTypeSubKey, contentType, 0)
	r.client.Set(key+":"+redisLengthSubKey, contentLength, 0)
	if data, err := ioutil.ReadAll(reader); err != nil {
		return err
	} else {
		if err := r.client.Set(key, string(data), 0).Err(); err != nil {
			return err
		}
	}
	return nil
}
func (r *RedisStorage) Delete(token string, filename string) error {
	key := formRedisKey(token, filename)
	return r.client.Del(key, key+":"+redisTypeSubKey, key+":"+redisLengthSubKey).Err()
}
func (r *RedisStorage) IsNotExist(err error) bool {
	if err == redis.Nil {
		return true
	}
	return false
}

func (r *RedisStorage) Type() string {
	return "redis"
}

func formRedisKey(token, filename string) string {
	return fmt.Sprintf("storage:%s:%s", token, filename)
}
