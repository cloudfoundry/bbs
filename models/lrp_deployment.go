package models

import (
	"errors"
	"time"
)

type LRPDeploymentFilter struct {
	DefinitionIds []string
}

func (d *LRPDeploymentCreation) DesiredLRPKey() DesiredLRPKey {
	return NewDesiredLRPKey(d.Definition.DefinitionId, d.Domain, d.Definition.LogGuid)
}

func (lrp *LRPDeploymentCreation) DesiredLRPRunInfo(createdAt time.Time) DesiredLRPRunInfo {
	d := lrp.Definition
	return d.DesiredLRPRunInfo(lrp.DesiredLRPKey(), createdAt)
}

func (lrp *LRPDeploymentCreation) LRPDeployment(modTag *ModificationTag) *LRPDeployment {
	lrp.Definition.DefinitionId = lrp.DefinitionId
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

func NewLRPDefinitionSchedulingInfo(
	definitionId string,
	logGuid string,
	memoryMb int32,
	diskMb int32,
	rootFs string,
	maxPids int32,
	volumePlacement *VolumePlacement,
	placementTags []string,
) *LRPDefinitionSchedulingInfo {
	return &LRPDefinitionSchedulingInfo{
		DefinitionId:    definitionId,
		LogGuid:         logGuid,
		MemoryMb:        memoryMb,
		DiskMb:          diskMb,
		RootFs:          rootFs,
		MaxPids:         maxPids,
		VolumePlacement: volumePlacement,
		PlacementTags:   placementTags,
	}
}

func NewLRPDeploymentSchedulingInfo(
	processGuid string,
	domain string,
	instances int32,
	annotation string,
	modTag ModificationTag,
	routes *Routes,
	definitions map[string]*LRPDefinitionSchedulingInfo,
) LRPDeploymentSchedulingInfo {
	return LRPDeploymentSchedulingInfo{
		ProcessGuid:     processGuid,
		Domain:          domain,
		Instances:       instances,
		Annotation:      annotation,
		ModificationTag: modTag,
		Routes:          routes,
		Definitions:     definitions,
	}
}

func (d *LRPDefinition) LRPDefinitionSchedulingInfo() *LRPDefinitionSchedulingInfo {
	return &LRPDefinitionSchedulingInfo{
		DefinitionId:    d.DefinitionId,
		LogGuid:         d.LogGuid,
		MemoryMb:        d.MemoryMb,
		DiskMb:          d.DiskMb,
		RootFs:          d.RootFs,
		MaxPids:         d.MaxPids,
		VolumePlacement: d.VolumePlacement,
		PlacementTags:   d.PlacementTags,
	}
}

func (d *LRPDeployment) LRPDeploymentSchedulingInfo() *LRPDeploymentSchedulingInfo {
	definitions := make(map[string]*LRPDefinitionSchedulingInfo)
	for defId, definition := range d.Definitions {
		definitions[defId] = definition.LRPDefinitionSchedulingInfo()
	}
	return &LRPDeploymentSchedulingInfo{
		ProcessGuid:     d.ProcessGuid,
		Domain:          d.Domain,
		Instances:       d.Instances,
		Annotation:      d.Annotation,
		ModificationTag: *d.ModificationTag,
		Routes:          d.Routes,
		Definitions:     definitions,
	}
}

func (d *LRPDeployment) DesiredLRP(definitionId string) (DesiredLRP, error) {
	definition, ok := d.Definitions[definitionId]
	if !ok {
		return DesiredLRP{}, errors.New("invalid-definition-id")
	}

	lrpKey := NewDesiredLRPKey(definition.DefinitionId, d.Domain, definition.LogGuid)
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
