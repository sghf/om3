package cmd

func init() {
	kind := "vol"

	cmdObject := newCmdVol()
	cmdObjectCollector := newCmdObjectCollector(kind)
	cmdObjectCollectorTag := newCmdObjectCollectorTag(kind)
	cmdObjectEdit := newCmdObjectEdit(kind)
	cmdObjectInstance := newCmdObjectInstance(kind)
	cmdObjectInstanceConfig := newCmdObjectInstanceConfig(kind)
	cmdObjectInstanceMonitor := newCmdObjectInstanceMonitor(kind)
	cmdObjectInstanceStatus := newCmdObjectInstanceStatus(kind)
	cmdObjectSet := newCmdObjectSet(kind)
	cmdObjectPrint := newCmdObjectPrint(kind)
	cmdObjectPrintConfig := newCmdObjectPrintConfig(kind)
	cmdObjectPush := newCmdObjectPush(kind)
	cmdObjectResource := newCmdObjectResource(kind)
	cmdObjectResourceConfig := newCmdObjectResourceConfig(kind)
	cmdObjectResourceMonitor := newCmdObjectResourceMonitor(kind)
	cmdObjectResourceStatus := newCmdObjectResourceStatus(kind)
	cmdObjectSync := newCmdObjectSync(kind)
	cmdObjectValidate := newCmdObjectValidate(kind)

	root.AddCommand(
		cmdObject,
	)
	cmdObject.AddCommand(
		cmdObjectCollector,
		cmdObjectEdit,
		cmdObjectInstance,
		cmdObjectPrint,
		cmdObjectPush,
		cmdObjectResource,
		cmdObjectSet,
		cmdObjectSync,
		cmdObjectValidate,
		newCmdObjectAbort(kind),
		newCmdObjectBoot(kind),
		newCmdObjectClear(kind),
		newCmdObjectCreate(kind),
		newCmdObjectDelete(kind),
		newCmdObjectDoc(kind),
		newCmdObjectEval(kind),
		newCmdObjectEnter(kind),
		newCmdObjectFreeze(kind),
		newCmdObjectGet(kind),
		newCmdObjectGiveback(kind),
		newCmdObjectLogs(kind),
		newCmdObjectLs(kind),
		newCmdObjectMonitor(kind),
		newCmdObjectPurge(kind),
		newCmdObjectProvision(kind),
		newCmdObjectPRStart(kind),
		newCmdObjectPRStop(kind),
		newCmdObjectRestart(kind),
		newCmdObjectRun(kind),
		newCmdObjectShutdown(kind),
		newCmdObjectStart(kind),
		newCmdObjectStartStandby(kind),
		newCmdObjectStatus(kind),
		newCmdObjectStop(kind),
		newCmdObjectSwitch(kind),
		newCmdObjectTakeover(kind),
		newCmdObjectThaw(kind),
		newCmdObjectUnfreeze(kind),
		newCmdObjectUnprovision(kind),
		newCmdObjectUnset(kind),
	)
	cmdObjectCollector.AddCommand(
		cmdObjectCollectorTag,
	)
	cmdObjectCollectorTag.AddCommand(
		newCmdObjectCollectorTagAttach(kind),
		newCmdObjectCollectorTagCreate(kind),
		newCmdObjectCollectorTagDetach(kind),
		newCmdObjectCollectorTagList(kind),
		newCmdObjectCollectorTagShow(kind),
	)
	cmdObjectEdit.AddCommand(
		newCmdObjectEditConfig(kind),
	)
	cmdObjectInstanceConfig.AddCommand(
		newCmdObjectInstanceConfigLs(kind),
	)
	cmdObjectInstanceMonitor.AddCommand(
		newCmdObjectInstanceMonitorLs(kind),
	)
	cmdObjectInstanceStatus.AddCommand(
		newCmdObjectInstanceStatusLs(kind),
	)
	cmdObjectInstance.AddCommand(
		cmdObjectInstanceConfig,
		cmdObjectInstanceMonitor,
		cmdObjectInstanceStatus,
		newCmdObjectInstanceLs(kind),
	)
	cmdObjectResourceConfig.AddCommand(
		newCmdObjectResourceConfigLs(kind),
	)
	cmdObjectResourceMonitor.AddCommand(
		newCmdObjectResourceMonitorLs(kind),
	)
	cmdObjectResourceStatus.AddCommand(
		newCmdObjectResourceStatusLs(kind),
	)
	cmdObjectResource.AddCommand(
		cmdObjectResourceConfig,
		cmdObjectResourceMonitor,
		cmdObjectResourceStatus,
		newCmdObjectResourceLs(kind),
	)
	cmdObjectSet.AddCommand(
		newCmdObjectSetProvisioned(kind),
		newCmdObjectSetUnprovisioned(kind),
	)
	cmdObjectPrint.AddCommand(
		cmdObjectPrintConfig,
		newCmdObjectPrintDevices(kind),
		newCmdObjectPrintSchedule(kind),
		newCmdObjectPrintStatus(kind),
	)
	cmdObjectPrintConfig.AddCommand(
		newCmdObjectPrintConfigMtime(kind),
	)
	cmdObjectPush.AddCommand(
		newCmdObjectPushResInfo(kind),
	)
	cmdObjectSync.AddCommand(
		newCmdObjectSyncFull(kind),
		newCmdObjectSyncResync(kind),
		newCmdObjectSyncUpdate(kind),
	)
	cmdObjectValidate.AddCommand(
		newCmdObjectValidateConfig(kind),
	)
}
