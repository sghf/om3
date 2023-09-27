package api

func (t Instance) Unstructured() map[string]any {
	m := map[string]any{}
	if t.Config != nil {
		m["config"] = t.Config.Unstructured()
	}
	if t.Monitor != nil {
		m["monitor"] = t.Monitor.Unstructured()
	}
	if t.Status != nil {
		m["status"] = t.Status.Unstructured()
	}
	return m
}

func (t InstanceMap) Unstructured() map[string]any {
	m := make(map[string]any)
	for k, v := range t {
		m[k] = v.Unstructured()
	}
	return m
}

func (t InstanceMeta) Unstructured() map[string]any {
	return map[string]any{
		"node":   t.Node,
		"object": t.Object,
	}
}

func (t InstanceItem) Unstructured() map[string]any {
	return map[string]any{
		"meta": t.Meta.Unstructured(),
		"data": t.Data.Unstructured(),
	}
}

func (t ObjectItem) Unstructured() map[string]any {
	return map[string]any{
		"meta": t.Meta.Unstructured(),
		"data": t.Data.Unstructured(),
	}
}

func (t ObjectMeta) Unstructured() map[string]any {
	return map[string]any{
		"object": t.Object,
	}
}

func (t ObjectData) Unstructured() map[string]any {
	m := map[string]any{
		"avail":              t.Avail,
		"flex_max":           t.FlexMax,
		"flex_min":           t.FlexMin,
		"flex_target":        t.FlexTarget,
		"frozen":             t.Frozen,
		"instances":          t.Instances.Unstructured(),
		"orchestrate":        t.Orchestrate,
		"overall":            t.Overall,
		"placement_policy":   t.PlacementPolicy,
		"placement_state":    t.PlacementState,
		"priority":           t.Priority,
		"provisioned":        t.Provisioned,
		"scope":              t.Scope,
		"topology":           t.Topology,
		"up_instances_count": t.UpInstancesCount,
		"updated_at":         t.UpdatedAt,
	}
	if t.Pool != nil {
		m["pool"] = *t.Pool
	}
	if t.Size != nil {
		m["size"] = *t.Size
	}
	return m
}

func (t Resource) Unstructured() map[string]any {
	return map[string]any{
		"config":  t.Config.Unstructured(),
		"monitor": t.Monitor.Unstructured(),
		"status":  t.Status.Unstructured(),
	}
}

func (t ResourceMeta) Unstructured() map[string]any {
	return map[string]any{
		"node":   t.Node,
		"object": t.Object,
		"rid":    t.Rid,
	}
}

func (t ResourceItem) Unstructured() map[string]any {
	return map[string]any{
		"meta": t.Meta.Unstructured(),
		"data": t.Data.Unstructured(),
	}
}

func (t NetworkIp) Unstructured() map[string]any {
	return map[string]any{
		"ip":      t.Ip,
		"network": t.Network,
		"node":    t.Node,
		"path":    t.Path,
		"rid":     t.Rid,
	}
}

func (t Network) Unstructured() map[string]any {
	return map[string]any{
		"errors":  t.Errors,
		"name":    t.Name,
		"network": t.Network,
		"free":    t.Free,
		"size":    t.Size,
		"type":    t.Type,
		"used":    t.Used,
	}
}

func (t Pool) Unstructured() map[string]any {
	return map[string]any{
		"type":         t.Type,
		"name":         t.Name,
		"capabilities": t.Capabilities,
		"head":         t.Head,
		"errors":       t.Errors,
		"volume_count": t.VolumeCount,
		"free":         t.Free,
		"used":         t.Used,
		"size":         t.Size,
	}
}

func (t PoolVolume) Unstructured() map[string]any {
	return map[string]any{
		"pool":      t.Pool,
		"path":      t.Path,
		"children":  t.Children,
		"is_orphan": t.IsOrphan,
		"size":      t.Size,
	}
}