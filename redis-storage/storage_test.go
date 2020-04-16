package redis_storage_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	redisStorage "github.com/dutchcoders/transfer.sh/redis-storage"
)

// TODO: add container for test

func TestRedisStorage(t *testing.T) {
	token := "test-token"
	filename := "example.md"
	content := "This is redis storage"
	storage := redisStorage.New("localhost:6379", "")

	r := bytes.NewReader([]byte(content))
	size := uint64(r.Size())

	err := storage.Put(token, filename, r, "text", size)
	assert.NoError(t, err)

	rc, s, err := storage.Get(token, filename)
	result, err := ioutil.ReadAll(rc)
	assert.NoError(t, err)

	assert.Equal(t, content, string(result))
	assert.Equal(t, size, s)
}
