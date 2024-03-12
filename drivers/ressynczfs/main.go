package ressynczfs

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/opensvc/om3/core/actioncontext"
	"github.com/opensvc/om3/core/driver"
	"github.com/opensvc/om3/core/nodesinfo"
	"github.com/opensvc/om3/core/provisioned"
	"github.com/opensvc/om3/core/rawconfig"
	"github.com/opensvc/om3/core/resource"
	"github.com/opensvc/om3/core/status"
	"github.com/opensvc/om3/core/statusbus"
	"github.com/opensvc/om3/core/topology"
	"github.com/opensvc/om3/drivers/ressync"
	"github.com/opensvc/om3/util/hostname"
	"github.com/opensvc/om3/util/proc"
	"github.com/opensvc/om3/util/progress"
	"github.com/opensvc/om3/util/sshnode"
	"github.com/opensvc/om3/util/zfs"
	"golang.org/x/crypto/ssh"
)

// T is the driver structure.
type (
	T struct {
		ressync.T
		Src       string
		Dst       string
		Target    []string
		Schedule  string
		Recursive bool
		Nodes     []string
		DRPNodes  []string
		ObjectID  uuid.UUID
		Timeout   *time.Duration
		Topology  topology.T
		User      string

		srcSnapSent   string
		srcSnapTosend string
		dstSnapSent   string
		dstSnapTosend string
	}

	modeT uint
)

const (
	modeFull modeT = iota
	modeIncr

	lockName = "sync"
)

func New() resource.Driver {
	return &T{}
}

func (t T) IsRunning() bool {
	unlock, err := t.Lock(false, time.Second*0, lockName)
	if err != nil {
		return true
	}
	defer unlock()
	return false
}

func (t T) Full(ctx context.Context) error {
	disable := actioncontext.IsLockDisabled(ctx)
	timeout := actioncontext.LockTimeout(ctx)
	target := actioncontext.Target(ctx)
	cancel, err := t.Lock(disable, timeout, lockName)
	if err != nil {
		return err
	}
	defer cancel()
	return t.lockedSync(ctx, modeFull, target)
}

func (t T) Update(ctx context.Context) error {
	disable := actioncontext.IsLockDisabled(ctx)
	timeout := actioncontext.LockTimeout(ctx)
	target := actioncontext.Target(ctx)
	cancel, err := t.Lock(disable, timeout, lockName)
	if err != nil {
		return err
	}
	defer cancel()
	return t.lockedSync(ctx, modeIncr, target)
}

func (t T) lockedSync(ctx context.Context, mode modeT, target []string) (err error) {
	if len(target) == 0 {
		target = t.Target
	}

	isCron := actioncontext.IsCron(ctx)

	if t.isFlexAndNotPrimary() {
		return fmt.Errorf("this flex instance is not primary. only %s can sync", t.Nodes[0])
	}

	if v, rids := t.isInstanceSufficientlyStarted(ctx); !v {
		return fmt.Errorf("the instance is not sufficiently started (%s). refuse to sync to protect the data of the started remote instance", strings.Join(rids, ","))
	}

	hasSnapSent, err := t.snapshotExists(t.srcSnapSent)

	if err != nil {
		return err
	}

	hasSnapTosend, err := t.snapshotExists(t.srcSnapTosend)

	if err != nil {
		return err
	}

	if !hasSnapSent {
		t.Log().Infof("%s does not exist: can't send delta, send full", t.srcSnapSent)
		mode = modeFull
		if err := t.zfs(t.srcSnapTosend).Destroy(zfs.FilesystemDestroyWithRecurse(t.Recursive)); err != nil {
			return err
		}
	} else if mode == modeFull {
		if err := t.zfs(t.srcSnapSent).Destroy(zfs.FilesystemDestroyWithRecurse(t.Recursive)); err != nil {
			return err
		}
		if err := t.zfs(t.srcSnapTosend).Destroy(zfs.FilesystemDestroyWithRecurse(t.Recursive)); err != nil {
			return err
		}
		if err := t.zfs(t.srcSnapTosend).Snapshot(zfs.FilesystemSnapshotWithRecursive(t.Recursive)); err != nil {
			return err
		}
	} else if !hasSnapTosend {
		if err := t.zfs(t.srcSnapTosend).Snapshot(zfs.FilesystemSnapshotWithRecursive(t.Recursive)); err != nil {
			return err
		}
	}

	nodenames := t.GetTargetPeernames(target, t.Nodes, t.DRPNodes)
	for _, nodename := range nodenames {
		if err := t.isSendAllowedToPeerEnv(nodename); err != nil {
			if isCron {
				t.Log().Debugf("%s", err)
			} else {
				t.Log().Infof("%s", err)
			}
			continue
		}
		t.progress(ctx, nodename)
		if mode == modeFull {
			if err := t.zfs(t.dstSnapSent).Destroy(zfs.FilesystemDestroyWithRecurse(t.Recursive), zfs.FilesystemDestroyWithNode(nodename)); err != nil {
				return err
			}
		}
		if err := t.zfs(t.dstSnapTosend).Destroy(zfs.FilesystemDestroyWithRecurse(t.Recursive), zfs.FilesystemDestroyWithNode(nodename)); err != nil {
			return err
		}
		if err := t.peerSync(ctx, mode, nodename); err != nil {
			return err
		}
		if err := t.rotatePeerSnaps(nodename, t.dstSnapTosend, t.srcSnapSent); err != nil {
			return err
		}
		if t.WritePeerLastSync(nodename, nodenames); err != nil {
			return err
		}
	}
	if err := t.rotateSnaps(t.srcSnapTosend, t.srcSnapSent); err != nil {
		return err
	}
	return nil
}

func (t *T) sendIncremental(nodename string) error {
	var b bytes.Buffer

	args := t.sendIncrementalCmd()
	cmd := exec.Command(args[0], args[1:]...)

	client, err := sshnode.NewClient(nodename)
	if err != nil {
		return err
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	if stdinPipe, err := session.StdinPipe(); err != nil {
		return err
	} else {
		cmd.Stdout = stdinPipe
		defer stdinPipe.Close()
	}

	session.Stdout = &b
	session.Stderr = &b

	rargs := t.receiveCmd()
	rcmd := exec.Command(rargs[0], rargs[1:]...)
	rcmdStr := rcmd.String()
	cmdStr := cmd.String()
	t.Log().Attr("cmd", fmt.Sprintf("%s | ssh %s '%s'", cmdStr, nodename, rcmdStr)).Infof("%s send delta to node %s", t.Src, nodename)
	if err := session.Start(rcmdStr); err != nil {
		ee := err.(*ssh.ExitError)
		ec := ee.Waitmsg.ExitStatus()
		t.Log().
			Attr("exitcode", ec).
			Attr("cmd", rcmdStr).
			Attr("host", nodename).
			Errorf("rexec '%s' on host %s exited with code %d: %s", rcmdStr, nodename, ec, string(b.Bytes()))
		return err
	}
	cmd.Stderr = &b
	err = cmd.Run()
	if ee, ok := err.(*exec.ExitError); ok {
		ec := ee.ExitCode()
		t.Log().
			Attr("exitcode", ec).
			Attr("cmd", cmdStr).
			Errorf("exec '%s' on host %s exited with code %d: %s", cmdStr, nodename, ec, string(b.Bytes()))
	}
	return err
}

func (t *T) sendInitial(nodename string) error {
	var b bytes.Buffer

	args := t.sendInitialCmd()
	cmd := exec.Command(args[0], args[1:]...)

	client, err := sshnode.NewClient(nodename)
	if err != nil {
		return err
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	if stdinPipe, err := session.StdinPipe(); err != nil {
		return err
	} else {
		cmd.Stdout = stdinPipe
		defer stdinPipe.Close()
	}

	session.Stdout = &b
	session.Stderr = &b

	rargs := t.receiveCmd()
	rcmd := exec.Command(rargs[0], rargs[1:]...)
	rcmdStr := rcmd.String()
	cmdStr := cmd.String()
	t.Log().Attr("cmd", fmt.Sprintf("%s | ssh %s '%s'", cmdStr, nodename, rcmdStr)).Infof("%s send full to node %s", t.Src, nodename)
	if err := session.Start(rcmdStr); err != nil {
		ee := err.(*ssh.ExitError)
		ec := ee.Waitmsg.ExitStatus()
		t.Log().
			Attr("exitcode", ec).
			Attr("cmd", rcmdStr).
			Attr("host", nodename).
			Errorf("rexec '%s' on host %s exited with code %d", rcmdStr, nodename, ec)
		return err
	}
	cmd.Stderr = &b
	err = cmd.Run()
	if ee, ok := err.(*exec.ExitError); ok {
		ec := ee.ExitCode()
		t.Log().
			Attr("exitcode", ec).
			Attr("cmd", cmd.String()).
			Errorf("exec '%s' on host %s exited with code %d: %s", cmdStr, nodename, ec, string(b.Bytes()))
	} else if err != nil {
		return err
	}
	return nil
}

func (t *T) sendInitialCmd() []string {
	cmd := []string{"zfs", "send"}
	if t.Recursive {
		cmd = append(cmd, "-R")
	} else {
		cmd = append(cmd, "-p")
	}
	cmd = append(cmd, t.srcSnapTosend)
	return cmd
}

func (t *T) sendIncrementalCmd() []string {
	cmd := []string{"zfs", "send"}
	if t.Recursive {
		cmd = append(cmd, "-R")
	}
	cmd = append(cmd, "-I", t.srcSnapSent, t.srcSnapTosend)
	return cmd
}

func getUpperFs(s string) string {
	l := filepath.SplitList(s)
	if len(l) > 1 {
		l = l[:len(l)-1]
		return filepath.Join(l...)
	} else {
		return "/"
	}
}

func (t *T) receiveCmd() []string {
	cmd := []string{"zfs", "receive"}
	srcPool := t.zfs(t.Src).PoolName()
	dstPool := t.zfs(t.Dst).PoolName()
	if t.Src == t.Dst || (t.Src == srcPool && t.Dst == dstPool) {
		cmd = append(cmd, "-dF", dstPool)
	} else {
		upperFs := getUpperFs(t.Dst)
		cmd = append(cmd, "-eF", upperFs)
	}
	return cmd
}

func (t *T) Kill(ctx context.Context) error {
	return nil
}

func (t *T) progress(ctx context.Context, nodename string, more ...any) {
	if view := progress.ViewFromContext(ctx); view != nil {
		key := append(t.ProgressKey(), nodename)
		view.Info(key, more)
	}
}

func (t *T) Status(ctx context.Context) status.T {
	var isSourceNode bool
	if v, _ := t.isInstanceSufficientlyStarted(ctx); !v {
		isSourceNode = false
	} else if t.isFlexAndNotPrimary() {
		isSourceNode = false
	} else {
		isSourceNode = true
	}
	nodenames := t.getTargetNodenames(isSourceNode)
	return t.StatusLastSync(nodenames)
}

func (t *T) running(ctx context.Context) bool {
	return false
}

// Label returns a formatted short description of the Resource
func (t T) Label() string {
	switch {
	case t.Src != "" && len(t.Target) > 0:
		return t.Src + " to " + strings.Join(t.Target, " ")
	case t.Src != "":
		return t.Src + " to void"
	case len(t.Target) > 0:
		return "nothing to " + strings.Join(t.Target, " ")
	default:
		return ""
	}
}

func (t T) getRunning(cmdArgs []string) (proc.L, error) {
	procs, err := proc.All()
	if err != nil {
		return procs, err
	}
	procs = procs.FilterByEnv("OPENSVC_ID", t.ObjectID.String())
	procs = procs.FilterByEnv("OPENSVC_RID", t.RID())
	return procs, nil
}

func (t T) ScheduleOptions() resource.ScheduleOptions {
	return resource.ScheduleOptions{
		Action: "sync_update",
		Option: "schedule",
		Base:   "",
	}
}

func (t T) Provisioned() (provisioned.T, error) {
	return provisioned.NotApplicable, nil
}

func (t *T) Configure() error {
	rid := strings.Replace(t.RID(), "#", ".", 1)
	t.srcSnapSent = t.Src + "@" + rid + ".sent"
	t.srcSnapTosend = t.Src + "@" + rid + ".tosend"
	t.dstSnapSent = t.Dst + "@" + rid + ".sent"
	t.dstSnapTosend = t.Dst + "@" + rid + ".tosend"
	return nil
}

func (t *T) zfs(name string) *zfs.Filesystem {
	return &zfs.Filesystem{Name: name, Log: t.Log()}
}

func (t *T) rotatePeerSnaps(nodename, src, dst string) error {
	if err := t.zfs(dst).Destroy(zfs.FilesystemDestroyWithRecurse(t.Recursive), zfs.FilesystemDestroyWithNode(nodename)); err != nil {
		return err
	}
	if err := t.zfs(src).Rename(dst, zfs.FilesystemRenameWithRecurse(t.Recursive), zfs.FilesystemRenameWithNode(nodename)); err != nil {
		return err
	}
	return nil
}

func (t *T) rotateSnaps(src, dst string) error {
	if err := t.zfs(dst).Destroy(zfs.FilesystemDestroyWithRecurse(t.Recursive)); err != nil {
		return err
	}
	if err := t.zfs(src).Rename(dst, zfs.FilesystemRenameWithRecurse(t.Recursive)); err != nil {
		return err
	}
	return nil
}

func (t *T) snapshotExists(name string) (bool, error) {
	return t.zfs(name).SnapshotExists()
}

func (t *T) remoteSnapshotExists(name, nodename string) (bool, error) {
	client, err := sshnode.NewClient(nodename)
	if err != nil {
		return false, err
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		return false, err
	}
	defer session.Close()

	cmd := fmt.Sprintf("zfs list -t snapshot %s", name)

	if b, err := session.CombinedOutput(cmd); err != nil {
		ee := err.(*ssh.ExitError)
		ec := ee.Waitmsg.ExitStatus()
		if ec == 0 {
			return true, nil
		}
		if strings.Contains(string(b), "does not exist") {
			return false, nil
		}
		t.Log().
			Attr("exitcode", ec).
			Attr("cmd", cmd).
			Attr("host", nodename).
			Debugf("rexec '%s' on host %s exited with code %d", cmd, nodename, ec)
		return false, err
	}
	return true, nil
}

func (t T) peerSync(ctx context.Context, mode modeT, nodename string) error {
	err := func() error {
		if mode == modeFull {
			return t.sendInitial(nodename)
		} else if v, err := t.remoteSnapshotExists(t.dstSnapSent, nodename); err != nil {
			return err
		} else if v {
			return t.sendIncremental(nodename)
		} else {
			return t.sendInitial(nodename)
		}
	}()

	var icon string
	if err != nil {
		icon = rawconfig.Colorize.Error("✓")
	} else {
		icon = rawconfig.Colorize.Optimal("✓")
	}
	t.progress(ctx, nodename, icon, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (t T) user() string {
	if t.User != "" {
		return t.User
	} else {
		return "root"
	}
}

func (t T) Info(ctx context.Context) (resource.InfoKeys, error) {
	target := sort.StringSlice(t.Target)
	sort.Sort(target)
	m := resource.InfoKeys{
		{Key: "src", Value: t.Src},
		{Key: "dst", Value: t.Dst},
		{Key: "recursive", Value: fmt.Sprintf("%v", t.Recursive)},
		{Key: "target", Value: strings.Join(target, " ")},
	}
	if t.Timeout != nil {
		m = append(m, resource.InfoKey{Key: "timeout", Value: fmt.Sprintf("%s", t.Timeout)})
	}
	return m, nil
}

func (t *T) isInstanceSufficientlyStarted(ctx context.Context) (v bool, rids []string) {
	sb := statusbus.FromContext(ctx)
	o := t.GetObjectDriver()
	l := o.ResourcesByDrivergroups([]driver.Group{
		driver.GroupIP,
		driver.GroupFS,
		driver.GroupShare,
		driver.GroupDisk,
		driver.GroupContainer,
	})
	v = true
	for _, r := range l {
		switch r.ID().DriverGroup() {
		case driver.GroupIP:
		case driver.GroupFS:
		case driver.GroupShare:
		case driver.GroupDisk:
			switch r.Manifest().DriverID.Name {
			case "drbd":
				continue
			case "scsireserv":
				continue
			}
		case driver.GroupContainer:
		default:
			continue
		}
		st := sb.Get(r.RID())
		switch st {
		case status.Up:
		case status.StandbyUp:
		case status.NotApplicable:
		default:
			// required resource is not up
			rids = append(rids, r.RID())
			v = false
		}
	}
	return
}

func (t *T) isFlexAndNotPrimary() bool {
	if t.Topology != topology.Flex {
		return false
	}
	if hostname.Hostname() == t.Nodes[0] {
		return false
	}
	return true
}

func (t *T) isSendAllowedToPeerEnv(nodename string) error {
	var localEnv, peerEnv string
	nodesInfo, err := nodesinfo.Load()
	if err != nil {
		return fmt.Errorf("get nodes info: %w", err)
	}
	getEnv := func(n string, s *string) error {
		if m, ok := nodesInfo[n]; !ok {
			return fmt.Errorf("node %s not found in nodes_info.json", n)
		} else {
			*s = m.Env
		}
		return nil
	}
	if err := getEnv(hostname.Hostname(), &localEnv); err != nil {
		return err
	}
	if err := getEnv(nodename, &peerEnv); err != nil {
		return err
	}
	if localEnv != "PRD" && peerEnv == "PRD" {
		return fmt.Errorf("refuse to sync from a non-PRD node to a PRD node")
	}
	return nil
}

func (t *T) getTargetNodenames(isSourceNode bool) []string {
	if isSourceNode {
		// if the instance is active, check last sync timestamp for each peer
		return t.GetTargetPeernames(t.Target, t.Nodes, t.DRPNodes)
	} else {
		// if the instance is passive, check last sync timestamp for the local node (received from the source node)
		return []string{hostname.Hostname()}
	}
}
