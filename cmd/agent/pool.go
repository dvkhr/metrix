package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

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

func (cw *CollectWorker) Run() {
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

type SendFunc func(ctx context.Context, mStor storage.MemStorage, cl *http.Client, serverAddress string, signKey []byte) error

func Retry(sendMetrics SendFunc, retries int) SendFunc {
	return func(ctx context.Context, mStor storage.MemStorage, cl *http.Client, serverAddress string, signKey []byte) error {
		for r := 0; ; r++ {
			nextAttemptAfter := time.Duration(2*r+1) * time.Second
			err := sendMetrics(ctx, mStor, cl, serverAddress, signKey)
			if err == nil || r >= retries {
				return err
			}
			fmt.Printf("Attempt %d failed; retrying in %v\n", r+1, nextAttemptAfter)
			select {
			case <-time.After(nextAttemptAfter):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

type SendWorker struct {
	wf            SendFunc
	poll          int64
	ctx           context.Context
	payloadChan   chan service.Metrics
	stopChan      chan bool
	cl            *http.Client
	mStor         storage.MemStorage
	serverAddress string
	signKey       []byte
	mtx           sync.Mutex
}

func (sw *SendWorker) Run() {
	var sendInterval time.Time
	sw.mStor.NewStorage()

	for {
		if sendInterval.IsZero() ||
			time.Since(sendInterval) >= time.Duration(sw.poll)*time.Second {
			sw.mtx.Lock()
			r := Retry(sw.wf, 3)
			err := r(sw.ctx, sw.mStor, sw.cl, sw.serverAddress, sw.signKey)
			if err != nil {
				fmt.Println(err)
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
			continue
		case <-sw.stopChan:
			return
		default:
			continue
		}
	}
}
