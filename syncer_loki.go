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
type LokiSyncerOption func(syncer *LokiSyncer)

// WithLokiAddr provide loki address
func WithLokiAddr(addr string) LokiSyncerOption {
	return func(syncer *LokiSyncer) {
		if len(addr) > 0 {
			syncer.addr = addr
		}
	}
}

// WithLokiPath provide loki path
func WithLokiPath(in string) LokiSyncerOption {
	return func(syncer *LokiSyncer) {
		if len(in) > 0 {
			syncer.path = in
		}
	}
}

// WithLokiUsername provide loki username
func WithLokiUsername(name string) LokiSyncerOption {
	return func(syncer *LokiSyncer) {
		syncer.username = name
	}
}

// WithLokiPassword provide loki password
func WithLokiPassword(pass string) LokiSyncerOption {
	return func(syncer *LokiSyncer) {
		syncer.password = pass
	}
}

// WithLokiClientTls provide loki http client TLS config
func WithLokiClientTls(conf *tls.Config) LokiSyncerOption {
	return func(syncer *LokiSyncer) {
		syncer.tlsConfig = conf
	}
}

// WithLokiLabel provide labels, should follow isValidLabelName()
func WithLokiLabel(key, value string) LokiSyncerOption {
	return func(syncer *LokiSyncer) {
		if len(key) > 0 && len(value) > 0 && isValidLabelName(key) {
			syncer.labels.Set(key, value)
		}
	}
}

// WithLokiMaxBatchWaitMs provide max batch wait time in milli
func WithLokiMaxBatchWaitMs(in time.Duration) LokiSyncerOption {
	return func(syncer *LokiSyncer) {
		if in.Milliseconds() > 0 {
			syncer.maxBatchWaitMs = in
		}
	}
}

// WithLokiMaxBatchSize provide max batch size
func WithLokiMaxBatchSize(batchSize int) LokiSyncerOption {
	return func(syncer *LokiSyncer) {
		if batchSize > 0 {
			syncer.maxBatchSize = batchSize
		}
	}
}

// NewLokiSyncer create new lokiSyncer
func NewLokiSyncer(opts ...LokiSyncerOption) *LokiSyncer {
	syncer := &LokiSyncer{
		addr:           "localhost:3100",
		path:           "/loki/api/v1/push",
		labels:         newAtomicMap(),
		maxBatchWaitMs: 3000 * time.Millisecond,
		maxBatchSize:   1000,
		quitChannel:    make(chan struct{}),
		logChannel:     make(chan *lokiValue),
	}

	for i := range opts {
		opts[i](syncer)
	}

	// convert label key if illegal
	syncer.labels.Set("rk_logger", "v1")

	// init http client
	syncer.initHttpClient()

	// init basic auth
	syncer.initBasicAuth()

	// add wait group
	syncer.waitGroup.Add(1)

	return syncer
}

// Init http client
func (syncer *LokiSyncer) initHttpClient() {
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
func (syncer *LokiSyncer) initBasicAuth() {
	if len(syncer.username) > 0 && len(syncer.password) > 0 {
		auth := syncer.username + ":" + syncer.password
		syncer.basicAuthHeader = "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	}
}

// LokiSyncer which will periodically send logs to Loki
type LokiSyncer struct {
	addr            string          `yaml:"addr" json:"addr"`
	path            string          `yaml:"path" json:"path"`
	username        string          `yaml:"username" json:"username"`
	password        string          `yaml:"-" json:"-"`
	basicAuthHeader string          `yaml:"-" json:"-"`
	tlsConfig       *tls.Config     `yaml:"-" json:"-"`
	maxBatchWaitMs  time.Duration   `yaml:"maxBatchWaitMs" json:"maxBatchWaitMs"`
	maxBatchSize    int             `yaml:"maxBatchSize" json:"maxBatchSize"`
	labels          *atomicMap      `yaml:"-" json:"-"`
	logChannel      chan *lokiValue `yaml:"-" json:"-"`
	quitChannel     chan struct{}   `yaml:"-" json:"-"`
	waitGroup       sync.WaitGroup  `yaml:"-" json:"-"`
	httpClient      *http.Client    `yaml:"-" json:"-"`
}

// Send message to remote loki server
func (syncer *LokiSyncer) send(entries []*lokiValue) {
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

// ************* Bootstrap & Interrupt *************

// Bootstrap run periodic jobs
func (syncer *LokiSyncer) Bootstrap(context.Context) {
	go func() {
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
	}()
}

// Interrupt goroutine
func (syncer *LokiSyncer) Interrupt(context.Context) {
	close(syncer.quitChannel)
	syncer.waitGroup.Wait()
}

// ************* Model *************
func (syncer *LokiSyncer) AddLabel(key, value string) {
	syncer.labels.Set(key, value)
}

// Create new lokiStreamList
func (syncer *LokiSyncer) newLokiStreamList(values []*lokiValue) []byte {
	msg := &lokiStreamList{
		Streams: []*lokiStream{},
	}

	for i := range values {
		val := values[i]
		labels := syncer.labels.Copy()
		for k, v := range val.Labels {
			labels[k] = v
		}

		msg.Streams = append(msg.Streams, &lokiStream{
			Stream: labels,
			Values: [][]string{val.Values},
		})
	}

	bytes, _ := json.Marshal(msg)

	return bytes
}

// Refer https://grafana.com/docs/loki/latest/api/#post-lokiapiv1push
type lokiValue struct {
	Values []string          `json:"-"`
	Labels map[string]string `json:"-"`
}

// Refer https://grafana.com/docs/loki/latest/api/#post-lokiapiv1push
type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// Refer https://grafana.com/docs/loki/latest/api/#post-lokiapiv1push
type lokiStreamList struct {
	Streams []*lokiStream `json:"streams"`
}

// ************* Implementation of zapcore.WriteSyncer *************

// Write to logChannel
func (syncer *LokiSyncer) Write(p []byte) (n int, err error) {
	syncer.logChannel <- &lokiValue{
		Values: []string{fmt.Sprintf("%d", time.Now().UnixNano()), string(p)},
	}

	return len(p), nil
}

// Noop
func (syncer *LokiSyncer) Sync() error {
	return nil
}

func newAtomicMap() *atomicMap {
	return &atomicMap{
		mutex: sync.Mutex{},
		m:     make(map[string]string),
	}
}

type atomicMap struct {
	mutex sync.Mutex
	m     map[string]string
}

func (a *atomicMap) Set(k, v string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.m[k] = v
}

func (a *atomicMap) Get(k string) string {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.m[k]
}

func (a *atomicMap) Copy() map[string]string {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	res := map[string]string{}
	for k, v := range a.m {
		res[k] = v
	}

	return res
}

func (a *atomicMap) Delete(k string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	delete(a.m, k)
}
