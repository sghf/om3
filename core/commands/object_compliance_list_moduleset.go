package commands

import (
	"github.com/spf13/cobra"
	"opensvc.com/opensvc/core/flag"
	"opensvc.com/opensvc/core/object"
	"opensvc.com/opensvc/core/objectaction"
	"opensvc.com/opensvc/core/path"
)

type (
	// CmdObjectComplianceListModuleset is the cobra flag set of the sysreport command.
	CmdObjectComplianceListModuleset struct {
		object.OptsObjectComplianceListModuleset
	}
)

// Init configures a cobra command and adds it to the parent command.
func (t *CmdObjectComplianceListModuleset) Init(kind string, parent *cobra.Command, selector *string) {
	cmd := t.cmd(kind, selector)
	parent.AddCommand(cmd)
	flag.Install(cmd, t)
}

func (t *CmdObjectComplianceListModuleset) cmd(kind string, selector *string) *cobra.Command {
	return &cobra.Command{
		Use:     "moduleset",
		Short:   "List compliance moduleset available to this node.",
		Aliases: []string{"modulese"},
		Run: func(_ *cobra.Command, _ []string) {
			t.run(selector, kind)
		},
	}
}

func (t *CmdObjectComplianceListModuleset) run(selector *string, kind string) {
	mergedSelector := mergeSelector(*selector, t.Global.ObjectSelector, kind, "")
	objectaction.New(
		objectaction.LocalFirst(),
		objectaction.WithLocal(t.Global.Local),
		objectaction.WithColor(t.Global.Color),
		objectaction.WithFormat(t.Global.Format),
		objectaction.WithObjectSelector(mergedSelector),
		objectaction.WithRemoteNodes(t.Global.NodeSelector),
		objectaction.WithServer(t.Global.Server),
		objectaction.WithRemoteAction("compliance list moduleset"),
		objectaction.WithRemoteOptions(map[string]interface{}{
			"format": t.Global.Format,
		}),
		objectaction.WithLocalRun(func(p path.T) (interface{}, error) {
			if o, err := object.NewSvc(p); err != nil {
				return nil, err
			} else {
				return o.ComplianceListModuleset(t.OptsObjectComplianceListModuleset)
			}
		}),
	).Do()
}