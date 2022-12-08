package instance

import (
	"encoding/json"
	"sort"
	"time"

	"opensvc.com/opensvc/core/kind"
	"opensvc.com/opensvc/core/path"
	"opensvc.com/opensvc/core/placement"
	"opensvc.com/opensvc/core/priority"
	"opensvc.com/opensvc/core/provisioned"
	"opensvc.com/opensvc/core/resource"
	"opensvc.com/opensvc/core/resourceid"
	"opensvc.com/opensvc/core/status"
	"opensvc.com/opensvc/core/topology"
	"opensvc.com/opensvc/util/stringslice"
)

type (
	Instance struct {
		Config  *Config  `json:"config"`
		Monitor *Monitor `json:"monitor"`
		Status  *Status  `json:"status"`
	}

	// Monitor describes the in-daemon states of an instance
	Monitor struct {
		GlobalExpect        string                    `json:"global_expect"`
		LocalExpect         string                    `json:"local_expect"`
		Status              string                    `json:"status"`
		StatusUpdated       time.Time                 `json:"status_updated"`
		GlobalExpectUpdated time.Time                 `json:"global_expect_updated"`
		LocalExpectUpdated  time.Time                 `json:"local_expect_updated"`
		IsLeader            bool                      `json:"is_leader"`
		IsHALeader          bool                      `json:"is_ha_leader"`
		Restart             map[string]MonitorRestart `json:"restart,omitempty"`
	}

	// MonitorRestart describes the restart states maintained by the daemon
	// for an object instance.
	MonitorRestart struct {
		Retries int       `json:"retries"`
		Updated time.Time `json:"updated"`
	}

	// Config describes a configuration file content checksum,
	// timestamp of last change and the nodes it should be installed on.
	Config struct {
		Checksum        string           `json:"csum"`
		FlexTarget      int              `json:"flex_target,omitempty"`
		FlexMin         int              `json:"flex_min,omitempty"`
		FlexMax         int              `json:"flex_max,omitempty"`
		Nodename        string           `json:"-"`
		Orchestrate     string           `json:"orchestrate"`
		Path            path.T           `json:"-"`
		PlacementPolicy placement.Policy `json:"placement_policy"`
		Priority        priority.T       `json:"priority,omitempty"`
		Scope           []string         `json:"scope"`
		Topology        topology.T       `json:"topology"`
		Updated         time.Time        `json:"updated"`
	}

	// Status describes the instance status.
	Status struct {
		App         string                   `json:"app,omitempty"`
		Avail       status.T                 `json:"avail"`
		Constraints bool                     `json:"constraints,omitempty"`
		DRP         bool                     `json:"drp,omitempty"`
		Overall     status.T                 `json:"overall"`
		Csum        string                   `json:"csum,omitempty"`
		Env         string                   `json:"env,omitempty"`
		Frozen      time.Time                `json:"frozen,omitempty"`
		Kind        kind.T                   `json:"kind"`
		Optional    status.T                 `json:"optional,omitempty"`
		Provisioned provisioned.T            `json:"provisioned"`
		Preserved   bool                     `json:"preserved,omitempty"`
		Updated     time.Time                `json:"updated"`
		Subsets     map[string]SubsetStatus  `json:"subsets,omitempty"`
		Resources   []resource.ExposedStatus `json:"resources,omitempty"`
		Running     ResourceRunningSet       `json:"running,omitempty"`
		Parents     []path.Relation          `json:"parents,omitempty"`
		Children    []path.Relation          `json:"children,omitempty"`
		Slaves      []path.Relation          `json:"slaves,omitempty"`
		StatusGroup map[string]string        `json:"status_group,omitempty"`
	}

	// ResourceOrder is a sortable list representation of the
	// instance status resources map.
	ResourceOrder []resource.ExposedStatus

	// ResourceRunningSet is the list of resource currently running (sync and task).
	ResourceRunningSet []string

	// SubsetStatus describes a resource subset properties.
	SubsetStatus struct {
		Parallel bool `json:"parallel,omitempty"`
	}
)

// Has is true if the rid is found running in the Instance Monitor data sent by the daemon.
func (t ResourceRunningSet) Has(rid string) bool {
	for _, r := range t {
		if r == rid {
			return true
		}
	}
	return false
}

// UnmarshalJSON serializes the type instance as JSON.
func (t *MonitorRestart) UnmarshalJSON(b []byte) error {
	type tempT MonitorRestart
	temp := tempT(MonitorRestart{})
	if err := json.Unmarshal(b, &temp); err != nil {
		var retries int
		if err := json.Unmarshal(b, &retries); err != nil {
			return err
		}
		temp.Retries = retries
	}
	*t = MonitorRestart(temp)
	return nil
}

// SortedResources returns a list of resource identifiers sorted by:
// 1/ driver group
// 2/ subset
// 3/ resource name
func (t *Status) SortedResources() []resource.ExposedStatus {
	l := make([]resource.ExposedStatus, 0)
	for _, v := range t.Resources {
		rid, err := resourceid.Parse(v.Rid)
		if err != nil {
			continue
		}
		v.ResourceID = rid
		l = append(l, v)
	}
	sort.Sort(ResourceOrder(l))
	return l
}

func (t Status) DeepCopy() *Status {
	t.Running = append(ResourceRunningSet{}, t.Running...)
	t.Parents = append([]path.Relation{}, t.Parents...)
	t.Children = append([]path.Relation{}, t.Children...)
	t.Slaves = append([]path.Relation{}, t.Slaves...)

	subSets := make(map[string]SubsetStatus)

	for id, v := range t.Subsets {
		subSets[id] = v
	}
	t.Subsets = subSets

	resources := make([]resource.ExposedStatus, 0)
	for _, v := range t.Resources {
		resources = append(resources, *v.DeepCopy())
	}
	t.Resources = resources

	statusGroup := make(map[string]string)
	for id, v := range t.StatusGroup {
		statusGroup[id] = v
	}
	t.StatusGroup = statusGroup

	return &t
}

func (a ResourceOrder) Len() int      { return len(a) }
func (a ResourceOrder) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ResourceOrder) Less(i, j int) bool {
	switch {
	case a[i].ResourceID.DriverGroup() < a[j].ResourceID.DriverGroup():
		return true
	case a[i].ResourceID.DriverGroup() > a[j].ResourceID.DriverGroup():
		return false
	// same driver group
	case a[i].Subset < a[j].Subset:
		return true
	case a[i].Subset > a[j].Subset:
		return false
	// and same subset
	default:
		return a[i].ResourceID.Name < a[j].ResourceID.Name
	}
}

// resourceFlagsString formats resource flags as a vector of characters.
//
//	R  Running
//	M  Monitored
//	D  Disabled
//	O  Optional
//	E  Encap
//	P  Provisioned
//	S  Standby
func (t Status) ResourceFlagsString(rid resourceid.T, r resource.ExposedStatus) string {
	flags := ""

	// Running task or sync
	if t.Running.Has(rid.Name) {
		flags += "R"
	} else {
		flags += "."
	}

	flags += r.Monitor.FlagString()
	flags += r.Disable.FlagString()
	flags += r.Optional.FlagString()
	flags += r.Encap.FlagString()
	flags += r.Provisioned.State.FlagString()
	flags += r.Standby.FlagString()
	return flags
}

func (smon Monitor) ResourceFlagRestartString(rid resourceid.T, r resource.ExposedStatus) string {
	// Restart and retries
	retries := 0
	if restart, ok := smon.Restart[rid.Name]; ok {
		retries = restart.Retries
	}
	return r.Restart.FlagString(retries)
}

func (cfg Config) DeepCopy() *Config {
	newCfg := cfg
	newCfg.Scope = append([]string{}, cfg.Scope...)
	return &newCfg
}

func (smon Monitor) DeepCopy() *Monitor {
	v := smon
	restart := make(map[string]MonitorRestart)
	for s, val := range v.Restart {
		restart[s] = val
	}
	v.Restart = restart
	return &v
}

// ConfigEqual returns a boolean reporting whether a == b
//
// Nodename and Path are not compared
func ConfigEqual(a, b *Config) bool {
	if a.Updated != b.Updated {
		return false
	}
	if a.Checksum != b.Checksum {
		return false
	}
	if !stringslice.Equal(a.Scope, b.Scope) {
		return false
	}
	return true
}
