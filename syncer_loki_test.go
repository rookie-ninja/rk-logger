package rklogger

import (
	"context"
	"crypto/tls"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func Test_IsValidLabelName(t *testing.T) {
	assert.False(t, isValidLabelName(""))
	assert.False(t, isValidLabelName("ut-key"))
	assert.True(t, isValidLabelName("ut_key"))
}

func TestNewLokiSyncer(t *testing.T) {
	// without options
	syncer := NewLokiSyncer()
	assert.Equal(t, "http://localhost:3100", syncer.addr)
	assert.Equal(t, "/loki/api/v1/push", syncer.path)
	assert.Empty(t, syncer.username)
	assert.Empty(t, syncer.password)
	assert.Empty(t, syncer.basicAuthHeader)
	assert.Nil(t, syncer.tlsConfig)
	assert.Equal(t, 3000*time.Millisecond, syncer.maxBatchWaitMs)
	assert.Equal(t, 1000, syncer.maxBatchSize)
	assert.NotNil(t, syncer.quitChannel)
	assert.NotNil(t, syncer.buffer)
	syncer.Bootstrap(context.TODO())
	syncer.Interrupt(context.TODO())

	// with options
	syncer = NewLokiSyncer(
		WithLokiAddr("ut-addr"),
		WithLokiPath("ut-path"),
		WithLokiUsername("ut-name"),
		WithLokiPassword("ut-pass"),
		WithLokiClientTls(&tls.Config{}),
		WithLokiLabel("key", "value"),
		WithLokiMaxBatchWaitMs(time.Second),
		WithLokiMaxBatchSize(10))
	assert.Equal(t, "https://ut-addr", syncer.addr)
	assert.Equal(t, "ut-path", syncer.path)
	assert.NotEmpty(t, syncer.basicAuthHeader)
	assert.NotNil(t, syncer.tlsConfig)
	assert.Equal(t, time.Second, syncer.maxBatchWaitMs)
	assert.Equal(t, 10, syncer.maxBatchSize)
	assert.NotNil(t, syncer.quitChannel)
	assert.NotNil(t, syncer.buffer)

	syncer.Bootstrap(context.TODO())
	syncer.Interrupt(context.TODO())
}

func TestLokiSyncer_send(t *testing.T) {
	defer assertNotPanic(t)

	syncer := LokiSyncer{
		httpClient:      http.DefaultClient,
		basicAuthHeader: "Basic xxx",
		buffer:          newAtomicSlice(),
	}

	syncer.send()
}

func TestLokiSyncer_Write(t *testing.T) {
	defer assertNotPanic(t)

	syncer := NewLokiSyncer()

	syncer.Write([]byte("ut"))
	assert.Equal(t, syncer.buffer.len(), 1)
}

func TestLokiSyncer_Sync(t *testing.T) {
	defer assertNotPanic(t)

	syncer := NewLokiSyncer()
	assert.Nil(t, syncer.Sync())
}

func TestAtomicMap(t *testing.T) {
	m := newAtomicMap()

	m.Set("key", "value")
	assert.Equal(t, "value", m.Get("key"))
	assert.Len(t, m.Copy(), 1)
	m.Delete("key")
	assert.Empty(t, m.Copy())
}

func assertNotPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, false)
	} else {
		// This should never be called in case of a bug
		assert.True(t, true)
	}
}
