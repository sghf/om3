package commands

import (
	"context"

	"github.com/opensvc/om3/core/naming"
	"github.com/opensvc/om3/core/object"
	"github.com/opensvc/om3/core/objectaction"
	"github.com/opensvc/om3/core/objectlogger"
)

type (
	CmdObjectComplianceAttachRuleset struct {
		OptsGlobal
		Ruleset string
	}
)

func (t *CmdObjectComplianceAttachRuleset) Run(selector, kind string) error {
	mergedSelector := mergeSelector(selector, t.ObjectSelector, kind, "")
	return objectaction.New(
		objectaction.LocalFirst(),
		objectaction.WithLocal(t.Local),
		objectaction.WithColor(t.Color),
		objectaction.WithOutput(t.Output),
		objectaction.WithObjectSelector(mergedSelector),
		objectaction.WithRemoteNodes(t.NodeSelector),
		objectaction.WithServer(t.Server),
		objectaction.WithRemoteAction("compliance attach ruleset"),
		objectaction.WithRemoteOptions(map[string]interface{}{
			"format":  t.Output,
			"ruleset": t.Ruleset,
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
				return nil, comp.AttachRuleset(t.Ruleset)
			}
		}),
	).Do()
}
