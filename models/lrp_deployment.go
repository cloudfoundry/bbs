package models

import (
	"errors"
	"time"
)

func (d *LRPDeploymentCreation) DesiredLRPKey() DesiredLRPKey {
	return NewDesiredLRPKey(d.ProcessGuid, d.Domain, d.Definition.LogGuid)
}

func (lrp *LRPDeploymentCreation) DesiredLRPRunInfo(createdAt time.Time) DesiredLRPRunInfo {
	d := lrp.Definition
	return d.DesiredLRPRunInfo(lrp.DesiredLRPKey(), createdAt)
}

func (lrp *LRPDeploymentCreation) LRPDeployment(modTag *ModificationTag) *LRPDeployment {
	return &LRPDeployment{
		ProcessGuid: lrp.ProcessGuid,
		Domain:      lrp.Domain,
		Instances:   lrp.Instances,
		Annotation:  lrp.Annotation,
		Routes:      lrp.Routes,
		Definitions: map[string]*LRPDefinition{
			lrp.DefinitionId: lrp.Definition,
		},
		ActiveDefinitionId: lrp.DefinitionId,
		ModificationTag:    modTag,
	}
}

func (d *LRPDeployment) DesiredLRP(definitionId string) (DesiredLRP, error) {
	definition, ok := d.Definitions[definitionId]
	if !ok {
		return DesiredLRP{}, errors.New("invalid-definition-id")
	}

	lrpKey := NewDesiredLRPKey(d.ProcessGuid, d.Domain, definition.LogGuid)
	runInfo := definition.DesiredLRPRunInfo(lrpKey, time.Now())
	schedInfo := NewDesiredLRPSchedulingInfo(
		lrpKey, d.Annotation, d.Instances, definition.DesiredLRPResource(),
		*d.Routes, *d.ModificationTag, definition.VolumePlacement, definition.PlacementTags)
	return NewDesiredLRP(schedInfo, runInfo), nil
}

func (d *LRPDefinition) DesiredLRPResource() DesiredLRPResource {
	return NewDesiredLRPResource(d.MemoryMb, d.DiskMb, d.MaxPids, d.RootFs)
}

func (d *LRPDefinition) DesiredLRPRunInfo(desiredLRPKey DesiredLRPKey, createdAt time.Time) DesiredLRPRunInfo {
	environmentVariables := make([]EnvironmentVariable, len(d.EnvironmentVariables))
	for i := range d.EnvironmentVariables {
		environmentVariables[i] = *d.EnvironmentVariables[i]
	}

	egressRules := make([]SecurityGroupRule, len(d.EgressRules))
	for i := range d.EgressRules {
		egressRules[i] = *d.EgressRules[i]
	}

	return NewDesiredLRPRunInfo(
		desiredLRPKey,
		createdAt,
		environmentVariables,
		d.CachedDependencies,
		d.Setup,
		d.Action,
		d.Monitor,
		d.StartTimeoutMs,
		d.Privileged,
		d.CpuWeight,
		d.Ports,
		egressRules,
		d.LogSource,
		d.MetricsGuid,
		d.LegacyDownloadUser,
		d.TrustedSystemCertificatesPath,
		d.VolumeMounts,
		d.Network,
		d.CertificateProperties,
		d.ImageUsername,
		d.ImagePassword,
		d.CheckDefinition,
	)
}
