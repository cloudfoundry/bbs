package models

import "time"

func (d *LRPDeploymentDefinition) DesiredLRPKey() DesiredLRPKey {
	return NewDesiredLRPKey(d.ProcessGuid, d.Domain, d.Definition.LogGuid)
}

func (lrp *LRPDeploymentDefinition) DesiredLRPRunInfo(createdAt time.Time) DesiredLRPRunInfo {
	d := lrp.Definition
	return d.DesiredLRPRunInfo(lrp.DesiredLRPKey(), createdAt)
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
