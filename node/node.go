package node

import (
	"context"
	"sync/atomic"
	"time"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/conflux-chain/conflux-infura/util"
	"github.com/sirupsen/logrus"
)

type Node struct {
	sdk.ClientOperator
	name         string
	cancel       context.CancelFunc
	atomicStatus atomic.Value
}

func NewNode(name, url string, hm HealthMonitor) *Node {
	ctx, cancel := context.WithCancel(context.Background())

	n := Node{
		ClientOperator: util.MustNewCfxClientWithRetry(url, 0, time.Millisecond),
		name:           name,
		cancel:         cancel,
	}

	n.atomicStatus.Store(NewStatus(name))

	go n.monitor(ctx, hm)

	return &n
}

func (n *Node) Name() string {
	return n.name
}

func (n *Node) Status() Status {
	return n.atomicStatus.Load().(Status)
}

func (n *Node) String() string {
	return n.name
}

func (n *Node) monitor(ctx context.Context, hm HealthMonitor) {
	ticker := time.NewTicker(cfg.Monitor.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logrus.WithField("name", n.name).Info("Complete to monitor node")
			return
		case <-ticker.C:
			status := n.atomicStatus.Load().(Status)
			status.Update(n, hm)
			n.atomicStatus.Store(status)
		}
	}
}

func (n *Node) Close() {
	n.ClientOperator.Close()
	n.cancel()
	status := n.Status()
	status.Close()
}