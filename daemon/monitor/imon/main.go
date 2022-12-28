// Package imon is responsible for of local instance state
//
//	It provides the cluster data:
//		["cluster", "node", <localhost>, "services", "status", <instance>, "monitor"]
//		["cluster", "node", <localhost>, "services", "imon", <instance>]
//
//	imon are created by the local instcfg, with parent context instcfg context.
//	instcfg done => imon done
//
//	worker watches on local instance status updates to clear reached status
//		=> unsetStatusWhenReached
//		=> orchestrate
//		=> pub new state if change
//
//	worker watches on remote instance imon updates converge global expects
//		=> convergeGlobalExpectFromRemote
//		=> orchestrate
//		=> pub new state if change
package imon

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"opensvc.com/opensvc/core/cluster"
	"opensvc.com/opensvc/core/instance"
	"opensvc.com/opensvc/core/object"
	"opensvc.com/opensvc/core/path"
	"opensvc.com/opensvc/daemon/daemondata"
	"opensvc.com/opensvc/daemon/msgbus"
	"opensvc.com/opensvc/util/hostname"
	"opensvc.com/opensvc/util/pubsub"
)

type (
	imon struct {
		state         instance.Monitor
		previousState instance.Monitor

		path     path.T
		id       string
		ctx      context.Context
		cancel   context.CancelFunc
		cmdC     chan any
		dataCmdC chan<- any
		log      zerolog.Logger

		pendingCtx    context.Context
		pendingCancel context.CancelFunc

		// updated data from object status update srcEvent
		instStatus  map[string]instance.Status
		instMonitor map[string]instance.Monitor
		nodeMonitor map[string]cluster.NodeMonitor
		nodeStats   map[string]cluster.NodeStats
		nodeStatus  map[string]cluster.NodeStatus
		scopeNodes  []string

		objStatus   object.Status
		cancelReady context.CancelFunc
		localhost   string
		change      bool

		sub *pubsub.Subscription
	}

	// cmdOrchestrate can be used from post action go routines
	cmdOrchestrate struct {
		state    instance.MonitorState
		newState instance.MonitorState
	}
)

// Start launch goroutine imon worker for a local instance state
func Start(parent context.Context, p path.T, nodes []string) error {
	ctx, cancel := context.WithCancel(parent)
	id := p.String()

	previousState := instance.Monitor{
		LocalExpect:  instance.MonitorLocalExpectUnset,
		GlobalExpect: instance.MonitorGlobalExpectUnset,
		State:        instance.MonitorStateIdle,
		Restart:      make(map[string]instance.MonitorRestart),
	}
	state := previousState

	o := &imon{
		state:         state,
		previousState: previousState,
		path:          p,
		id:            id,
		ctx:           ctx,
		cancel:        cancel,
		cmdC:          make(chan any),
		dataCmdC:      daemondata.BusFromContext(ctx),
		log:           log.Logger.With().Str("func", "imon").Stringer("object", p).Logger(),
		instStatus:    make(map[string]instance.Status),
		instMonitor:   make(map[string]instance.Monitor),
		nodeStatus:    make(map[string]cluster.NodeStatus),
		nodeStats:     make(map[string]cluster.NodeStats),
		nodeMonitor:   make(map[string]cluster.NodeMonitor),
		localhost:     hostname.Hostname(),
		scopeNodes:    nodes,
		change:        true,
	}

	o.startSubscriptions()

	go func() {
		defer func() {
			msgbus.DropPendingMsg(o.cmdC, time.Second)
			o.sub.Stop()
		}()
		o.worker(nodes)
	}()

	return nil
}

func (o *imon) startSubscriptions() {
	bus := pubsub.BusFromContext(o.ctx)
	sub := bus.Sub(o.id + "imon")
	label := pubsub.Label{"path", o.id}
	sub.AddFilter(msgbus.ObjectStatusUpdated{}, label)
	sub.AddFilter(msgbus.SetInstanceMonitor{}, label)
	sub.AddFilter(msgbus.InstanceMonitorUpdated{}, label)
	sub.AddFilter(msgbus.InstanceMonitorDeleted{}, label)
	sub.AddFilter(msgbus.NodeMonitorUpdated{})
	sub.AddFilter(msgbus.NodeStatusUpdated{})
	sub.AddFilter(msgbus.NodeStatsUpdated{})
	sub.Start()
	o.sub = sub
}

// worker watch for local imon updates
func (o *imon) worker(initialNodes []string) {
	defer o.log.Debug().Msg("done")

	for _, node := range initialNodes {
		o.instStatus[node] = daemondata.GetInstanceStatus(o.dataCmdC, o.path, node)
	}
	o.updateIfChange()
	defer o.delete()

	if err := o.crmStatus(); err != nil {
		o.log.Error().Err(err).Msg("error during initial crm status")
	}
	o.log.Debug().Msg("started")
	for {
		select {
		case <-o.ctx.Done():
			return
		case i := <-o.sub.C:
			switch c := i.(type) {
			case msgbus.ObjectStatusUpdated:
				o.onObjectStatusUpdated(c)
			case msgbus.SetInstanceMonitor:
				o.onSetInstanceMonitorClient(c.Monitor)
			case msgbus.InstanceMonitorUpdated:
				o.onInstanceMonitorUpdated(c)
			case msgbus.InstanceMonitorDeleted:
				o.onInstanceMonitorDeleted(c)
			case msgbus.NodeMonitorUpdated:
				o.onNodeMonitorUpdated(c)
			case msgbus.NodeStatusUpdated:
				o.onNodeStatusUpdated(c)
			case msgbus.NodeStatsUpdated:
				o.onNodeStatsUpdated(c)
			}
		case i := <-o.cmdC:
			switch c := i.(type) {
			case cmdOrchestrate:
				o.needOrchestrate(c)
			}
		}
	}
}

func (o *imon) delete() {
	if err := daemondata.DelInstanceMonitor(o.dataCmdC, o.path); err != nil {
		o.log.Error().Err(err).Msg("DelInstanceMonitor")
	}
}

func (o *imon) update() {
	newValue := o.state
	if err := daemondata.SetInstanceMonitor(o.dataCmdC, o.path, newValue); err != nil {
		o.log.Error().Err(err).Msg("SetInstanceMonitor")
	}
}

func (o *imon) transitionTo(newState instance.MonitorState) {
	o.change = true
	o.state.State = newState
	o.updateIfChange()
}

// updateIfChange log updates and publish new state value when changed
func (o *imon) updateIfChange() {
	if !o.change {
		return
	}
	o.change = false
	now := time.Now()
	previousVal := o.previousState
	newVal := o.state
	if newVal.GlobalExpect != previousVal.GlobalExpect {
		// Don't update GlobalExpectUpdated here
		// GlobalExpectUpdated is updated only during cmdSetInstanceMonitorClient and
		// its value is used for convergeGlobalExpectFromRemote
		o.loggerWithState().Info().Msgf("change monitor global expect %s -> %s", previousVal.GlobalExpect, newVal.GlobalExpect)
	}
	if newVal.LocalExpect != previousVal.LocalExpect {
		o.state.LocalExpectUpdated = now
		o.loggerWithState().Info().Msgf("change monitor local expect %s -> %s", previousVal.LocalExpect, newVal.LocalExpect)
	}
	if newVal.State != previousVal.State {
		o.state.StateUpdated = now
		o.loggerWithState().Info().Msgf("change monitor state %s -> %s", previousVal.State, newVal.State)
	}
	if newVal.IsLeader != previousVal.IsLeader {
		o.loggerWithState().Info().Msgf("change leader state %t -> %t", previousVal.IsLeader, newVal.IsLeader)
	}
	if newVal.IsHALeader != previousVal.IsHALeader {
		o.loggerWithState().Info().Msgf("change ha leader state %t -> %t", previousVal.IsHALeader, newVal.IsHALeader)
	}
	o.previousState = o.state
	o.update()
}

func (o *imon) hasOtherNodeActing() bool {
	for remoteNode, remoteInstMonitor := range o.instMonitor {
		if remoteNode == o.localhost {
			continue
		}
		if remoteInstMonitor.State.IsDoing() {
			return true
		}
	}
	return false
}

func (o *imon) createPendingWithCancel() {
	o.pendingCtx, o.pendingCancel = context.WithCancel(o.ctx)
}

func (o *imon) createPendingWithDuration(duration time.Duration) {
	o.pendingCtx, o.pendingCancel = context.WithTimeout(o.ctx, duration)
}

func (o *imon) clearPending() {
	if o.pendingCancel != nil {
		o.pendingCancel()
		o.pendingCancel = nil
		o.pendingCtx = nil
	}
}

func (o *imon) loggerWithState() *zerolog.Logger {
	ctx := o.log.With()
	if o.state.GlobalExpect != instance.MonitorGlobalExpectEmpty {
		ctx.Str("global_expect", o.state.GlobalExpect.String())
	}
	if o.state.LocalExpect != instance.MonitorLocalExpectEmpty {
		ctx.Str("local_expect", o.state.LocalExpect.String())
	}
	stateLogger := ctx.Logger()
	return &stateLogger
}