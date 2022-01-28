package rklogger

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// isValidLabelName returns true iff name qualified for loki label name
func isValidLabelName(name string) bool {
	if len(name) == 0 {
		return false
	}
	for i, b := range name {
		if !((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' || b == ':' || (b >= '0' && b <= '9' && i > 0)) {
			return false
		}
	}
	return true
}

// LokiSyncerOption options for lokiSyncer
type LokiSyncerOption func(syncer *lokiSyncer)

// WithLokiAddr provide loki address
func WithLokiAddr(addr string) LokiSyncerOption {
	return func(syncer *lokiSyncer) {
		if len(addr) > 0 {
			syncer.addr = addr
		}
	}
}

// WithLokiPath provide loki path
func WithLokiPath(in string) LokiSyncerOption {
	return func(syncer *lokiSyncer) {
		if len(in) > 0 {
			syncer.path = in
		}
	}
}

// WithLokiUsername provide loki username
func WithLokiUsername(name string) LokiSyncerOption {
	return func(syncer *lokiSyncer) {
		syncer.username = name
	}
}

// WithLokiPassword provide loki password
func WithLokiPassword(pass string) LokiSyncerOption {
	return func(syncer *lokiSyncer) {
		syncer.password = pass
	}
}

// WithClientTls provide loki http client TLS config
func WithClientTls(conf *tls.Config) LokiSyncerOption {
	return func(syncer *lokiSyncer) {
		syncer.tlsConfig = conf
	}
}

// WithLabel provide labels, should follow isValidLabelName()
func WithLabel(key, value string) LokiSyncerOption {
	return func(syncer *lokiSyncer) {
		if len(key) > 0 && len(value) > 0 {
			syncer.labels[key] = value
		}
	}
}

// WithMaxBatchWaitMs provide max batch wait time in milli
func WithMaxBatchWaitMs(in time.Duration) LokiSyncerOption {
	return func(syncer *lokiSyncer) {
		if in.Milliseconds() > 0 {
			syncer.maxBatchWaitMs = in
		}
	}
}

// WithMaxBatchSize provide max batch size
func WithMaxBatchSize(batchSize int) LokiSyncerOption {
	return func(syncer *lokiSyncer) {
		if batchSize > 0 {
			syncer.maxBatchSize = batchSize
		}
	}
}

// NewLokiSyncer create new lokiSyncer
func NewLokiSyncer(opts ...LokiSyncerOption) *lokiSyncer {
	syncer := &lokiSyncer{
		addr:           "localhost:3100",
		path:           "/loki/api/v1/push",
		labels:         make(map[string]string),
		maxBatchWaitMs: 3000 * time.Millisecond,
		maxBatchSize:   1000,
		quitChannel:    make(chan struct{}),
		logChannel:     make(chan *lokiValue),
	}

	for i := range opts {
		opts[i](syncer)
	}

	// convert label key if illegal
	syncer.labels["rk_logger"] = "v1"
	for k := range syncer.labels {
		if !isValidLabelName(k) {
			delete(syncer.labels, k)
		}
	}

	// init http client
	syncer.initHttpClient()

	// init basic auth
	syncer.initBasicAuth()

	// add wait group
	syncer.waitGroup.Add(1)

	go syncer.run()

	return syncer
}

// Init http client
func (syncer *lokiSyncer) initHttpClient() {
	// adjust loki addr
	strings.TrimPrefix(syncer.addr, "http://")
	strings.TrimPrefix(syncer.addr, "https://")

	syncer.httpClient = &http.Client{}

	if syncer.tlsConfig != nil {
		syncer.httpClient.Transport = &http.Transport{
			TLSClientConfig: syncer.tlsConfig,
		}

		syncer.addr = "https://" + syncer.addr
	} else {
		syncer.addr = "http://" + syncer.addr
	}
}

// Init basic auth header
func (syncer *lokiSyncer) initBasicAuth() {
	if len(syncer.username) > 0 && len(syncer.password) > 0 {
		auth := syncer.username + ":" + syncer.password
		syncer.basicAuthHeader = "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	}
}

// Loki syncer which will periodically send logs to Loki
type lokiSyncer struct {
	addr            string            `yaml:"addr" json:"addr"`
	path            string            `yaml:"path" json:"path"`
	username        string            `yaml:"username" json:"username"`
	password        string            `yaml:"password" json:"password"`
	basicAuthHeader string            `yaml:"-" json:"-"`
	tlsConfig       *tls.Config       `yaml:"-" json:"-"`
	maxBatchWaitMs  time.Duration     `yaml:"maxBatchWaitMs" json:"maxBatchWaitMs"`
	maxBatchSize    int               `yaml:"maxBatchSize" json:"maxBatchSize"`
	labels          map[string]string `yaml:"-" json:"-"`
	logChannel      chan *lokiValue   `yaml:"-" json:"-"`
	quitChannel     chan struct{}     `yaml:"-" json:"-"`
	waitGroup       sync.WaitGroup    `yaml:"-" json:"-"`
	httpClient      *http.Client      `yaml:"-" json:"-"`
}

// run periodic jobs
func (syncer *lokiSyncer) run() {
	var batch []*lokiValue
	batchSize := 0
	waitChannel := time.NewTimer(syncer.maxBatchWaitMs)

	defer func() {
		if batchSize > 0 {
			syncer.send(batch)
		}

		syncer.waitGroup.Done()
	}()

	for {
		select {
		case <-syncer.quitChannel:
			return
		case entry := <-syncer.logChannel:
			batch = append(batch, entry)
			batchSize++
			if batchSize >= syncer.maxBatchSize {
				syncer.send(batch)
				batch = []*lokiValue{}
				batchSize = 0
				waitChannel.Reset(syncer.maxBatchWaitMs)
			}
		case <-waitChannel.C:
			if batchSize > 0 {
				syncer.send(batch)
				batch = []*lokiValue{}
				batchSize = 0
			}
			waitChannel.Reset(syncer.maxBatchWaitMs)
		}
	}
}

// Send message to remote loki server
func (syncer *lokiSyncer) send(entries []*lokiValue) {
	streams := syncer.newLokiStreamList(entries)

	req, _ := http.NewRequest(http.MethodPost, syncer.addr+syncer.path, bytes.NewBuffer(streams))
	req.Header.Set("Content-Type", "application/json")
	if len(syncer.basicAuthHeader) > 0 {
		req.Header.Add("Authorization", syncer.basicAuthHeader)
	}

	resp, err := syncer.httpClient.Do(req)

	if err != nil {
		log.Printf("Failed to send an HTTP request: %s\n", err)
		return
	}

	if resp.StatusCode != 204 {
		log.Printf("Unexpected HTTP status code: %d\n", resp.StatusCode)
		return
	}
}

// ************* Interrupter *************

// Interrupt goroutine
func (syncer *lokiSyncer) Interrupt(context.Context) {
	close(syncer.quitChannel)
	syncer.waitGroup.Wait()
}

// ************* Model *************

// Create new lokiStreamList
func (syncer *lokiSyncer) newLokiStreamList(values []*lokiValue) []byte {
	msg := &lokiStreamList{
		Streams: []*lokiStream{
			{
				Stream: syncer.labels,
				Values: values,
			},
		},
	}

	bytes, _ := json.Marshal(msg)

	return bytes
}

// Refer https://grafana.com/docs/loki/latest/api/#post-lokiapiv1push
type lokiValue []string

// Refer https://grafana.com/docs/loki/latest/api/#post-lokiapiv1push
type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values []*lokiValue      `json:"values"`
}

// Refer https://grafana.com/docs/loki/latest/api/#post-lokiapiv1push
type lokiStreamList struct {
	Streams []*lokiStream `json:"streams"`
}

// ************* Implementation of zapcore.WriteSyncer *************

// Write to logChannel
func (syncer *lokiSyncer) Write(p []byte) (n int, err error) {
	syncer.logChannel <- &lokiValue{
		fmt.Sprintf("%d", time.Now().UnixNano()), string(p),
	}

	return len(p), nil
}

// Noop
func (syncer *lokiSyncer) Sync() error {
	return nil
}
