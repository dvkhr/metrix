package main

import (
	"context"
	"crypto/rsa"
	"net/http"
	"sync"
	"time"

	pb "github.com/dvkhr/metrix.git/internal/grpc/proto"
	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/retry"
	"github.com/dvkhr/metrix.git/internal/sender"
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

	for range pollTicker.C {
		cw.wf(cw.ctx, cw.payloadChan)
		select {
		case <-cw.stopChan:
			return
		default:
			continue
		}
	}

	defer pollTicker.Stop()
}

type SendWorker struct {
	mtx           sync.Mutex
	wfHTTP        retry.SendFunc
	wfGRPC        retry.SendFunc
	poll          int64
	ctx           context.Context
	payloadChan   chan service.Metrics
	stopChan      chan bool
	cl            *http.Client
	grpcClient    pb.MetricsServiceClient
	mStor         storage.MemStorage
	serverAddress string
	signKey       []byte
	publicKey     *rsa.PublicKey
	useGRPC       bool
}

func (sw *SendWorker) Run() {
	var sendInterval time.Time
	sw.mStor.NewStorage()

	for {
		if sendInterval.IsZero() ||
			time.Since(sendInterval) >= time.Duration(sw.poll)*time.Second {
			sw.mtx.Lock()

			options := sender.SendOptions{
				MemStorage:    sw.mStor,
				Client:        sw.cl,
				ServerAddress: sw.serverAddress,
				SignKey:       sw.signKey,
				PublicKey:     sw.publicKey,
			}

			r := retry.Retry(sw.wfHTTP, 3)
			err := r(sw.ctx, options)
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
		case <-sw.stopChan:
			return
		default:
			continue
		}
	}
}
