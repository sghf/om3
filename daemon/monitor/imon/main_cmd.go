package imon

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
	"time"

	"opensvc.com/opensvc/core/instance"
	"opensvc.com/opensvc/core/nodeselector"
	"opensvc.com/opensvc/core/placement"
	"opensvc.com/opensvc/core/provisioned"
	"opensvc.com/opensvc/core/status"
	"opensvc.com/opensvc/core/topology"
	"opensvc.com/opensvc/daemon/msgbus"
	"opensvc.com/opensvc/util/stringslice"
)

func (o *imon) onInstanceStatusUpdated(srcNode string, srcCmd msgbus.InstanceStatusUpdated) {
	if _, ok := o.instStatus[srcCmd.Node]; ok {
		if o.instStatus[srcCmd.Node].Updated.Before(srcCmd.Status.Updated) {
			// only update if more recent
			o.log.Debug().Msgf("ObjectStatusUpdated %s from InstanceStatusUpdated on %s update instance status", srcNode, srcCmd.Node)
			o.instStatus[srcCmd.Node] = srcCmd.Status
		} else {
			o.log.Debug().Msgf("ObjectStatusUpdated %s from InstanceStatusUpdated on %s skip update instance from obsolete status", srcNode, srcCmd.Node)
		}
	} else {
		o.log.Debug().Msgf("ObjectStatusUpdated %s from InstanceStatusUpdated on %s create instance status", srcNode, srcCmd.Node)
		o.instStatus[srcCmd.Node] = srcCmd.Status
	}
}

func (o *imon) onCfgUpdated(srcNode string, srcCmd msgbus.CfgUpdated) {
	if srcCmd.Node == o.localhost {
		cfgNodes := make(map[string]any)
		for _, node := range srcCmd.Config.Scope {
			cfgNodes[node] = nil
			if _, ok := o.instStatus[node]; !ok {
				o.instStatus[node] = instance.Status{Avail: status.Undef}
			}
		}
		for node := range o.instStatus {
			if _, ok := cfgNodes[node]; !ok {
				o.log.Info().Msgf("drop not anymore in local config status from node %s", node)
				delete(o.instStatus, node)
			}
		}
	}
	o.scopeNodes = append([]string{}, srcCmd.Config.Scope...)
	o.log.Debug().Msgf("updated from %s ObjectStatusUpdated CfgUpdated on %s scopeNodes=%s", srcNode, srcCmd.Node, o.scopeNodes)
}

func (o *imon) onCfgDeleted(srcNode string, srcCmd msgbus.CfgDeleted) {
	if _, ok := o.instStatus[srcCmd.Node]; ok {
		o.log.Info().Msgf("drop deleted instance status from node %s", srcCmd.Node)
		delete(o.instStatus, srcCmd.Node)
	}
}

// onObjectStatusUpdated updateIfChange state global expect from object status
func (o *imon) onObjectStatusUpdated(c msgbus.ObjectStatusUpdated) {
	if c.SrcEv != nil {
		switch srcCmd := c.SrcEv.(type) {
		case msgbus.InstanceStatusUpdated:
			o.onInstanceStatusUpdated(c.Node, srcCmd)
		case msgbus.CfgUpdated:
			o.onCfgUpdated(c.Node, srcCmd)
		case msgbus.CfgDeleted:
			o.onCfgDeleted(c.Node, srcCmd)
		}
	}
	o.objStatus = c.Status
	o.updateIsLeader()
	o.orchestrate()
}

func (o *imon) onSetInstanceMonitorClient(c instance.Monitor) {
	doStatus := func() {
		if _, ok := instance.MonitorStateStrings[c.State]; !ok {
			o.log.Warn().Msgf("invalid set instance monitor state: %s", c.State)
			return
		}
		if c.State == instance.MonitorStateEmpty {
			return
		}
		if o.state.State == c.State {
			o.log.Info().Msgf("instance monitor state is already %s", c.State)
			return
		}
		o.log.Info().Msgf("set instance monitor state %s -> %s", o.state.State, c.State)
		o.change = true
		o.state.State = c.State
	}

	doGlobalExpect := func() {
		if _, ok := instance.MonitorGlobalExpectStrings[c.GlobalExpect]; !ok {
			o.log.Warn().Msgf("invalid set instance monitor global expect: %s", c.GlobalExpect)
			return
		}
		switch c.GlobalExpect {
		case instance.MonitorGlobalExpectEmpty:
			return
		case instance.MonitorGlobalExpectPlacedAt:
			options, ok := c.GlobalExpectOptions.(instance.MonitorGlobalExpectOptionsPlacedAt)
			if !ok || len(options.Destination) == 0 {
				// Switch cmd without explicit target nodes.
				// Select some nodes automatically.
				dst := o.nextPlacedAtCandidate()
				if dst == "" {
					o.log.Info().Msg("no destination node could be selected from candidates")
					return
				}
				options.Destination = []string{dst}
				c.GlobalExpectOptions = options
			} else {
				want := options.Destination
				can := o.nextPlacedAtCandidates(want)
				if can == "" {
					o.log.Info().Msgf("no destination node could be selected from %s", want)
					return
				} else if can != want[0] {
					o.log.Info().Msgf("change destination nodes from %s to %s", want, can)
				}
				options.Destination = []string{can}
				c.GlobalExpectOptions = options
			}
		case instance.MonitorGlobalExpectStarted:
			if v, reason := o.isStartable(); !v {
				o.log.Info().Msg(reason)
				return
			}
		}
		for node, instMon := range o.instMonitor {
			if instMon.GlobalExpect == c.GlobalExpect {
				continue
			}
			if instMon.GlobalExpect == instance.MonitorGlobalExpectEmpty {
				continue
			}
			if instMon.GlobalExpectUpdated.After(o.state.GlobalExpectUpdated) {
				o.log.Info().Msgf("global expect is already %s on node %s", c.GlobalExpect, node)
				return
			}
		}

		if c.GlobalExpect != o.state.GlobalExpect {
			o.change = true
			o.state.GlobalExpect = c.GlobalExpect
			o.state.GlobalExpectOptions = c.GlobalExpectOptions
			// update GlobalExpectUpdated now
			// This will allow remote nodes to pickup most recent value
			o.state.GlobalExpectUpdated = time.Now()
		}
	}

	doLocalExpect := func() {
		switch c.LocalExpect {
		case instance.MonitorLocalExpectEmpty:
			return
		case instance.MonitorLocalExpectStarted:
		default:
			o.log.Warn().Msgf("invalid set instance monitor local expect: %s", c.LocalExpect)
			return
		}
		target := c.LocalExpect
		if o.state.LocalExpect == target {
			o.log.Info().Msgf("local expect is already %s", c.LocalExpect)
			return
		}
		o.log.Info().Msgf("set local expect %s -> %s", o.state.LocalExpect, target)
		o.change = true
		o.state.LocalExpect = target
	}

	doStatus()
	doGlobalExpect()
	doLocalExpect()

	if o.change {
		o.updateIfChange()
		o.orchestrate()
	}

}

func (o *imon) onNodeMonitorUpdated(c msgbus.NodeMonitorUpdated) {
	o.nodeMonitor[c.Node] = c.Monitor
	o.updateIsLeader()
	o.orchestrate()
	o.updateIfChange()
}

func (o *imon) onNodeStatusUpdated(c msgbus.NodeStatusUpdated) {
	o.nodeStatus[c.Node] = c.Value
	o.updateIsLeader()
	o.orchestrate()
	o.updateIfChange()
}

func (o *imon) onNodeStatsUpdated(c msgbus.NodeStatsUpdated) {
	o.nodeStats[c.Node] = c.Value
	if o.objStatus.PlacementPolicy == placement.Score {
		o.updateIsLeader()
		o.orchestrate()
		o.updateIfChange()
	}
}

func (o *imon) onInstanceMonitorUpdated(c msgbus.InstanceMonitorUpdated) {
	if c.Node != o.localhost {
		o.onRemoteInstanceMonitorUpdated(c)
	}
}

func (o *imon) onRemoteInstanceMonitorUpdated(c msgbus.InstanceMonitorUpdated) {
	remote := c.Node
	instMon := c.Status
	o.log.Debug().Msgf("updated instance imon from node %s  -> %s", remote, instMon.GlobalExpect)
	o.instMonitor[remote] = instMon
	o.convergeGlobalExpectFromRemote()
	o.updateIfChange()
	o.orchestrate()
	o.updateIfChange()
}

func (o *imon) onInstanceMonitorDeleted(c msgbus.InstanceMonitorDeleted) {
	node := c.Node
	if node == o.localhost {
		return
	}
	o.log.Debug().Msgf("delete remote instance imon from node %s", node)
	delete(o.instMonitor, c.Node)
	o.convergeGlobalExpectFromRemote()
	o.updateIfChange()
	o.orchestrate()
	o.updateIfChange()
}

func (o imon) GetInstanceMonitor(node string) (instance.Monitor, bool) {
	if o.localhost == node {
		return o.state, true
	}
	m, ok := o.instMonitor[node]
	return m, ok
}

func (o *imon) AllInstanceMonitorState(s instance.MonitorState) bool {
	for _, instMon := range o.AllInstanceMonitors() {
		if instMon.State != s {
			return false
		}
	}
	return true
}

func (o imon) AllInstanceMonitors() map[string]instance.Monitor {
	m := make(map[string]instance.Monitor)
	m[o.localhost] = o.state
	for node, instMon := range o.instMonitor {
		m[node] = instMon
	}
	return m
}

func (o imon) isExtraInstance() (bool, string) {
	if o.state.IsHALeader {
		return false, "object is not leader"
	}
	if v, reason := o.isHAOrchestrateable(); !v {
		return false, reason
	}
	if o.objStatus.Avail != status.Up {
		return false, "object is not up"
	}
	if o.objStatus.Topology != topology.Flex {
		return false, "object is not flex"
	}
	if o.objStatus.UpInstancesCount <= o.objStatus.FlexTarget {
		return false, fmt.Sprintf("%d/%d up instances", o.objStatus.UpInstancesCount, o.objStatus.FlexTarget)
	}
	return true, ""
}

func (o imon) isHAOrchestrateable() (bool, string) {
	if o.objStatus.Avail == status.Warn {
		return false, "object is warn state"
	}
	switch o.objStatus.Provisioned {
	case provisioned.Mixed:
		return false, "mixed object provisioned state"
	case provisioned.False:
		return false, "false object provisioned state"
	}
	return true, ""
}

func (o imon) isStartable() (bool, string) {
	if v, reason := o.isHAOrchestrateable(); !v {
		return false, reason
	}
	if o.isStarted() {
		return false, "already started"
	}
	return true, "object is startable"
}

func (o imon) isStarted() bool {
	switch o.objStatus.Topology {
	case topology.Flex:
		return o.objStatus.UpInstancesCount >= o.objStatus.FlexTarget
	case topology.Failover:
		return o.objStatus.Avail == status.Up
	default:
		return false
	}
}

func (o *imon) needOrchestrate(c cmdOrchestrate) {
	if c.state == instance.MonitorStateEmpty {
		return
	}
	if o.state.State == c.state {
		o.change = true
		o.state.State = c.newState
		o.updateIfChange()
	}
	o.orchestrate()
}

func (o *imon) sortCandidates(candidates []string) []string {
	switch o.objStatus.PlacementPolicy {
	case placement.NodesOrder:
		return o.sortWithNodesOrderPolicy(candidates)
	case placement.Spread:
		return o.sortWithSpreadPolicy(candidates)
	case placement.Score:
		return o.sortWithScorePolicy(candidates)
	case placement.Shift:
		return o.sortWithShiftPolicy(candidates)
	default:
		return []string{}
	}
}

func (o *imon) sortWithSpreadPolicy(candidates []string) []string {
	l := append([]string{}, candidates...)
	sum := func(s string) []byte {
		b := append([]byte(o.path.String()), []byte(s)...)
		return md5.New().Sum(b)
	}
	sort.SliceStable(l, func(i, j int) bool {
		return bytes.Compare(sum(l[i]), sum(l[j])) < 0
	})
	return l
}

// sortWithScorePolicy sorts candidates by descending cluster.NodeStats.Score
func (o *imon) sortWithScorePolicy(candidates []string) []string {
	l := append([]string{}, candidates...)
	sort.SliceStable(l, func(i, j int) bool {
		var si, sj uint64
		if stats, ok := o.nodeStats[l[i]]; ok {
			si = stats.Score
		}
		if stats, ok := o.nodeStats[l[j]]; ok {
			sj = stats.Score
		}
		return si > sj
	})
	return l
}

func (o *imon) sortWithLoadAvgPolicy(candidates []string) []string {
	o.log.Warn().Msg("TODO: sortWithLoadAvgPolicy")
	return candidates
}

func (o *imon) sortWithShiftPolicy(candidates []string) []string {
	var i int
	l := o.sortWithNodesOrderPolicy(candidates)
	l = append(l, l...)
	n := len(candidates)
	scalerSliceIndex := o.path.ScalerSliceIndex()
	if n > 0 && scalerSliceIndex > n {
		i = o.path.ScalerSliceIndex() % n
	}
	return candidates[i : i+n]
}

func (o *imon) sortWithNodesOrderPolicy(candidates []string) []string {
	var l []string
	for _, node := range o.scopeNodes {
		if stringslice.Has(node, candidates) {
			l = append(l, node)
		}
	}
	return l
}

func (o *imon) nextPlacedAtCandidates(want []string) string {
	expr := strings.Join(want, " ")
	var wantNodes []string
	for _, node := range nodeselector.LocalExpand(expr) {
		if _, ok := o.instStatus[node]; !ok {
			continue
		}
		wantNodes = append(wantNodes, node)
	}
	return strings.Join(wantNodes, ",")
}

func (o *imon) nextPlacedAtCandidate() string {
	if o.objStatus.Topology == topology.Flex {
		return ""
	}
	var candidates []string
	candidates = append(candidates, o.scopeNodes...)
	candidates = o.sortCandidates(candidates)

	for _, candidate := range candidates {
		if instStatus, ok := o.instStatus[candidate]; ok {
			switch instStatus.Avail {
			case status.Down, status.StandbyDown, status.StandbyUp:
				return candidate
			}
		}
	}
	return ""
}

func (o imon) IsInstanceStartFailed(node string) (bool, bool) {
	instMon, ok := o.GetInstanceMonitor(node)
	if !ok {
		return false, false
	}
	switch instMon.State {
	case instance.MonitorStateStartFailed:
		return true, true
	default:
		return false, true
	}
}

func (o imon) IsNodeMonitorStatusRankable(node string) (bool, bool) {
	nodeMonitor, ok := o.nodeMonitor[node]
	if !ok {
		return false, false
	}
	return nodeMonitor.State.IsRankable(), true
}

func (o *imon) newIsHALeader() bool {
	var candidates []string

	for _, node := range o.scopeNodes {
		if nodeStatus, ok := o.nodeStatus[node]; !ok || nodeStatus.IsFrozen() {
			continue
		}
		if instStatus, ok := o.instStatus[node]; !ok || instStatus.IsFrozen() {
			continue
		}
		if failed, ok := o.IsInstanceStartFailed(node); !ok || failed {
			continue
		}
		if v, ok := o.IsNodeMonitorStatusRankable(node); !ok || !v {
			continue
		}
		candidates = append(candidates, node)
	}
	candidates = o.sortCandidates(candidates)

	var maxLeaders int = 1
	if o.objStatus.Topology == topology.Flex {
		maxLeaders = o.objStatus.FlexTarget
	}

	i := stringslice.Index(o.localhost, candidates)
	if i < 0 {
		return false
	}
	return i < maxLeaders
	return false
}

func (o *imon) newIsLeader() bool {
	var candidates []string
	for _, node := range o.scopeNodes {
		if failed, ok := o.IsInstanceStartFailed(node); !ok || failed {
			continue
		}
		candidates = append(candidates, node)
	}
	candidates = o.sortCandidates(candidates)

	var maxLeaders int = 1
	if o.objStatus.Topology == topology.Flex {
		maxLeaders = o.objStatus.FlexTarget
	}

	i := stringslice.Index(o.localhost, candidates)
	if i < 0 {
		return false
	}
	return i < maxLeaders
}

func (o *imon) updateIsLeader() {
	if instStatus, ok := o.instStatus[o.localhost]; !ok || instStatus.Avail == status.NotApplicable {
		return
	}
	isLeader := o.newIsLeader()
	if isLeader != o.state.IsLeader {
		o.change = true
		o.state.IsLeader = isLeader
	}
	isHALeader := o.newIsHALeader()
	if isHALeader != o.state.IsHALeader {
		o.change = true
		o.state.IsHALeader = isHALeader
	}
	o.updateIfChange()
	return
}

// doTransitionAction execute action and update transition states
func (o *imon) doTransitionAction(action func() error, newState, successState, errorState instance.MonitorState) {
	o.transitionTo(newState)
	if action() != nil {
		o.transitionTo(errorState)
	} else {
		o.transitionTo(successState)
	}
}

// doAction runs action + background orchestration from action state result
//
// 1- set transient state to newState
// 2- run action
// 3- go orchestrateAfterAction(newState, successState or errorState)
func (o *imon) doAction(action func() error, newState, successState, errorState instance.MonitorState) {
	o.transitionTo(newState)
	nextState := successState
	if action() != nil {
		nextState = errorState
	}
	go o.orchestrateAfterAction(newState, nextState)
}