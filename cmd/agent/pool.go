package main

import (
	"context"
	"crypto/rsa"
	"net/http"
	"sync"
	"time"

	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/retry"
	"github.com/dvkhr/metrix.git/internal/service"
	"github.com/dvkhr/metrix.git/internal/storage"
)

type CollectFunc func(ctx context.Context, metrics chan service.Metrics)
type CollectWorker struct {
	wf          CollectFunc
	poll        int64
	ctx         context.Context
	payloadChan chan service.Metrics
	stopChan    chan bool
}

func (cw *CollectWorker) StartCollecting() {
	pollTicker := time.NewTicker(time.Duration(cw.poll) * time.Second)
	defer pollTicker.Stop()

	for {
		select {
		case <-cw.stopChan:
			logging.Logg.Info("Stopping CollectWorker due to stop signal")
			return
		case <-cw.ctx.Done():
			logging.Logg.Info("Stopping CollectWorker due to context cancellation")
			return
		case <-pollTicker.C:
			cw.wf(cw.ctx, cw.payloadChan)
		}
	}
}

type SendWorker struct {
	mtx           sync.Mutex
	wf            retry.SendFunc
	poll          int64
	ctx           context.Context
	payloadChan   chan service.Metrics
	stopChan      chan bool
	cl            *http.Client
	mStor         storage.MemStorage
	serverAddress string
	signKey       []byte
	publicKey     *rsa.PublicKey
}

func (sw *SendWorker) Run() {
	var sendInterval time.Time
	sw.mStor.NewStorage()

	for {
		select {
		case <-sw.stopChan:
			logging.Logg.Info("Stopping SendWorker due to stop signal")
			return
		case <-sw.ctx.Done():
			logging.Logg.Info("Stopping SendWorker due to context cancellation")
			return
		default:
			if sendInterval.IsZero() || time.Since(sendInterval) >= time.Duration(sw.poll)*time.Second {
				sw.mtx.Lock()
				r := retry.Retry(sw.wf, 3)
				err := r(sw.ctx, sw.mStor, sw.cl, sw.serverAddress, sw.signKey, sw.publicKey)
				if err != nil {
					logging.Logg.Error("Send worker error", "error", err)
				}
				sw.mStor.NewStorage()
				sendInterval = time.Now()
				sw.mtx.Unlock()
			}

			select {
			case mtrx := <-sw.payloadChan:
				sw.mtx.Lock()
				sw.mStor.Save(sw.ctx, mtrx)
				sw.mtx.Unlock()
			default:
				continue
			}
		}
	}
}
