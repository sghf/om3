package commands

import (
	"context"

	"github.com/opensvc/om3/core/naming"
	"github.com/opensvc/om3/core/object"
	"github.com/opensvc/om3/core/objectaction"
	"github.com/opensvc/om3/core/objectlogger"
)

type (
	CmdObjectComplianceListModules struct {
		OptsGlobal
	}
)

func (t *CmdObjectComplianceListModules) Run(selector, kind string) error {
	mergedSelector := mergeSelector(selector, t.ObjectSelector, kind, "")
	return objectaction.New(
		objectaction.LocalFirst(),
		objectaction.WithLocal(t.Local),
		objectaction.WithColor(t.Color),
		objectaction.WithOutput(t.Output),
		objectaction.WithObjectSelector(mergedSelector),
		objectaction.WithRemoteNodes(t.NodeSelector),
		objectaction.WithServer(t.Server),
		objectaction.WithRemoteAction("compliance env"),
		objectaction.WithRemoteOptions(map[string]interface{}{
			"format": t.Output,
		}),
		objectaction.WithLocalRun(func(ctx context.Context, p naming.Path) (interface{}, error) {
			logger := objectlogger.New(p,
				objectlogger.WithColor(t.Color != "no"),
				objectlogger.WithConsoleLog(t.Log != ""),
				objectlogger.WithLogFile(true),
			)
			if o, err := object.NewSvc(p, object.WithLogger(logger)); err != nil {
				return nil, err
			} else {
				comp, err := o.NewCompliance()
				if err != nil {
					return nil, err
				}
				return comp.ListModuleNames()
			}
		}),
	).Do()
}
