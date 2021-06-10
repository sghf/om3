package resappforking

import (
	"github.com/rs/zerolog"
	"opensvc.com/opensvc/core/resource"
	"opensvc.com/opensvc/core/status"
	"opensvc.com/opensvc/drivers/resapp"
	"opensvc.com/opensvc/util/xexec"
	"os/exec"
)

// T is the driver structure.
type T struct {
	resapp.T
}

func New() resource.Driver {
	return &T{}
}

func init() {
	resource.Register(driverGroup, driverName, New)
}

// Start the Resource
func (t T) Start() (err error) {
	t.Log().Debug().Msg("Start()")
	var xcmd xexec.T
	if xcmd, err = t.PrepareXcmd(t.StartCmd, "start"); err != nil {
		return
	} else if len(xcmd.CmdArgs) == 0 {
		return
	}
	appStatus := t.Status()
	if appStatus == status.Up {
		t.Log().Info().Msg("already up")
		return nil
	}
	command := exec.Command(xcmd.CmdArgs[0], xcmd.CmdArgs[1:]...)
	if err = xcmd.Update(command); err != nil {
		return
	}
	t.Log().Debug().Msg("Starting()")
	cmd := xexec.NewCmd(t.Log(), command, xexec.NewLoggerExec(t.Log(), zerolog.InfoLevel, zerolog.WarnLevel))
	if timeout := t.GetTimeout("start"); timeout > 0 {
		cmd.SetDuration(timeout)
	}
	t.Log().Info().Msgf("starting %s", command.String())
	// TODO Create PG
	if err = cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}

// Label returns a formatted short description of the Resource
func (t T) Label() string {
	return driverGroup.String()
}
