package hbctrl

import (
	"context"
	"time"

	"opensvc.com/opensvc/core/cluster"
	"opensvc.com/opensvc/daemon/daemonctx"
	"opensvc.com/opensvc/util/timestamp"
)

// peerWatch starts a new peer watcher of nodename for hbId
// when beating state change a hb_beating or hb_stale event is fired
// Once beating, a hb_stale event is fired if no beating are received after timeout
func (t *T) peerWatch(ctx context.Context, beatingC chan bool, hbId, nodename string, timeout time.Duration) {
	peer := cluster.HeartbeatPeerStatus{}
	beatingCtx, cancel := context.WithCancel(ctx)
	update := func(hbStatus cluster.HeartbeatPeerStatus) {
		t.cmd <- CmdSetPeerStatus{
			Nodename:   nodename,
			HbId:       hbId,
			PeerStatus: hbStatus,
		}
	}
	event := func(name string) {
		t.cmd <- CmdEvent{
			Name:     name,
			Nodename: nodename,
			HbId:     hbId,
		}
	}
	started := make(chan bool)
	go func() {
		defer cancel()
		log := daemonctx.Logger(ctx).With().Str("Name", "peerWatch-"+hbId+"-"+nodename).Logger()
		log.Info().Msg("watching")
		started <- true
		for {
			select {
			case <-t.ctx.Done():
				log.Info().Msg("done watching")
				return
			case beating := <-beatingC:
				if beating {
					if !peer.Beating {
						peer.Beating = true
						event("hb_beating")
					}
					cancel()
					beatingCtx, cancel = context.WithTimeout(t.ctx, timeout)
					peer.Last = timestamp.Now()
					update(peer)
				} else if peer.Beating {
					event("hb_stale")
				}
			case <-beatingCtx.Done():
				if peer.Beating {
					peer.Beating = false
					event("hb_stale")
					update(peer)
					beatingCtx, cancel = context.WithCancel(t.ctx)
				}
			}
		}
	}()
	<-started
}