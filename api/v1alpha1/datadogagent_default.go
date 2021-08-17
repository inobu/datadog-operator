// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package v1alpha1

import (
	"path"
	"strings"
	"time"

	"github.com/DataDog/datadog-operator/pkg/utils"
	edsdatadoghqv1alpha1 "github.com/DataDog/extendeddaemonset/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// default values
const (
	defaultLogLevel                         string = "INFO"
	defaultAgentImageTag                    string = "7.28.0"
	defaultClusterAgentImageTag             string = "1.12.0"
	defaultAgentImageName                   string = "agent"
	defaultClusterAgentImageName            string = "cluster-agent"
	defaultCollectEvents                    bool   = false
	defaultLeaderElection                   bool   = false
	defaultDockerSocketPath                 string = "/var/run/docker.sock"
	defaultDogstatsdOriginDetection         bool   = false
	defaultUseDogStatsDSocketVolume         bool   = false
	defaultHostDogstatsdSocketName          string = "statsd.sock"
	defaultHostDogstatsdSocketPath          string = "/var/run/datadog"
	defaultAgentEnabled                     bool   = true
	defaultClusterAgentEnabled              bool   = true
	defaultClusterChecksRunnerEnabled       bool   = false
	defaultApmEnabled                       bool   = false
	defaultApmHostPort                      int32  = 8126
	defaultExternalMetricsEnabled           bool   = false
	defaultSystemProbeEnabled               bool   = false
	defaultSystemProbeOOMKillEnabled        bool   = false
	defaultSystemProbeTCPQueueLengthEnabled bool   = false
	defaultSystemProbeConntrackEnabled      bool   = false
	defaultSystemProbeCollectDNSStats       bool   = false
	defaultSystemProbeBPFDebugEnabled       bool   = false
	defaultSystemProbeSecCompRootPath       string = "/var/lib/kubelet/seccomp"
	defaultAppArmorProfileName              string = "unconfined"
	DefaultSeccompProfileName               string = "localhost/system-probe"
	defaultSecurityRuntimeEnabled           bool   = false
	defaultSecurityComplianceEnabled        bool   = false
	defaultSecuritySyscallMonitorEnabled    bool   = false
	defaultHostApmSocketName                string = "apm.sock"
	defaultHostApmSocketPath                string = "/var/run/datadog"
	defaultLogEnabled                       bool   = false
	defaultLogsConfigContainerCollectAll    bool   = false
	defaultLogsContainerCollectUsingFiles   bool   = true
	defaultContainerLogsPath                string = "/var/lib/docker/containers"
	defaultPodLogsPath                      string = "/var/log/pods"
	defaultContainerSymlinksPath            string = "/var/log/containers"
	defaultLogsTempStoragePath              string = "/var/lib/datadog-agent/logs"
	defaultLogsOpenFilesLimit               int32  = 100
	defaultProcessEnabled                   bool   = false
	// `false` defaults to live container, agent activated but no process collection
	defaultProcessCollectionEnabled                      bool   = false
	defaultOrchestratorExplorerEnabled                   bool   = true
	defaultOrchestratorExplorerContainerScrubbingEnabled bool   = true
	DefaultOrchestratorExplorerConf                      string = "orchestrator-explorer-config"
	defaultMetricsProviderPort                           int32  = 8443
	defaultClusterChecksEnabled                          bool   = false
	DefaultKubeStateMetricsCoreConf                      string = "kube-state-metrics-core-config"
	defaultKubeStateMetricsCoreEnabled                   bool   = false
	defaultPrometheusScrapeEnabled                       bool   = false
	defaultPrometheusScrapeServiceEndpoints              bool   = false
	defaultRollingUpdateMaxUnavailable                          = "10%"
	defaultUpdateStrategy                                       = appsv1.RollingUpdateDaemonSetStrategyType
	defaultRollingUpdateMaxPodSchedulerFailure                  = "10%"
	defaultRollingUpdateMaxParallelPodCreation           int32  = 250
	defaultRollingUpdateSlowStartIntervalDuration               = 1 * time.Minute
	defaultRollingUpdateSlowStartAdditiveIncrease               = "5"
	defaultReconcileFrequency                                   = 10 * time.Second
	defaultRbacCreate                                           = true
	defaultMutateUnlabelled                                     = false
	DefaultAdmissionServiceName                                 = "datadog-admission-controller"
	defaultAdmissionControllerEnabled                           = false

	// Liveness probe default config
	defaultLivenessProbeInitialDelaySeconds int32 = 15
	defaultLivenessProbePeriodSeconds       int32 = 15
	defaultLivenessProbeTimeoutSeconds      int32 = 5
	defaultLivenessProbeSuccessThreshold    int32 = 1
	defaultLivenessProbeFailureThreshold    int32 = 6
	defaultAgentHealthPort                  int32 = 5555
	defaultLivenessProbeHTTPPath                  = "/live"

	// Readiness probe default config
	defaultReadinessProbeInitialDelaySeconds int32 = 15
	defaultReadinessProbePeriodSeconds       int32 = 15
	defaultReadinessProbeTimeoutSeconds      int32 = 5
	defaultReadinessProbeSuccessThreshold    int32 = 1
	defaultReadinessProbeFailureThreshold    int32 = 6
	defaultReadinessProbeHTTPPath                  = "/ready"
)

var defaultImagePullPolicy = corev1.PullIfNotPresent

// DefaultDatadogAgent defaults the DatadogAgent
func DefaultDatadogAgent(dda *DatadogAgent) *DatadogAgentStatus {
	// instOverrideStatus contains all the defaults from the runtime.
	// It is published in the status of the DatadogAgent
	dso := &DatadogAgentStatus{
		DefaultOverride: &DatadogAgentSpec{},
	}

	// Override spec given featureset
	FeatureOverride(&dda.Spec, dso.DefaultOverride)

	// Features
	// default features because it might have an impact on the other defaulting
	dso.DefaultOverride.Features = *DefaultFeatures(dda)

	// Cluster Agent
	dso.DefaultOverride.ClusterAgent = *DefaultDatadogAgentSpecClusterAgent(&dda.Spec.ClusterAgent)

	// Agent
	dso.DefaultOverride.Agent = *DefaultDatadogAgentSpecAgent(&dda.Spec.Agent)

	// CLC
	dso.DefaultOverride.ClusterChecksRunner = *DefaultDatadogAgentSpecClusterChecksRunner(&dda.Spec.ClusterChecksRunner)

	// Creds
	if dda.Spec.Credentials == nil {
		dda.Spec.Credentials = &AgentCredentials{}
	}
	if dda.Spec.Credentials.UseSecretBackend == nil {
		dda.Spec.Credentials.UseSecretBackend = NewBoolPointer(false)
		dso.DefaultOverride.Credentials = &AgentCredentials{
			UseSecretBackend: dda.Spec.Credentials.UseSecretBackend,
		}
	}

	return dso
}

// FeatureOverride defaults the feature section of the DatadogAgent
// TODO surface in the status when Overrides are not possible. Security agent requires the system probe
func FeatureOverride(dda *DatadogAgentSpec, dso *DatadogAgentSpec) {
	if dda.Features.NetworkMonitoring != nil && BoolValue(dda.Features.NetworkMonitoring.Enabled) {
		// If the Network monitoring Feature is enable, enable the System Probe.
		if !BoolValue(dda.Agent.Enabled) || dda.Agent.SystemProbe != nil {
			dda.Agent.SystemProbe.Enabled = NewBoolPointer(true)
			dso.Agent.SystemProbe = DefaultDatadogAgentSpecAgentSystemProbe(&dda.Agent)
			dso.Agent.SystemProbe.Enabled = NewBoolPointer(true)
		}
	}
	if dda.Features.OrchestratorExplorer != nil && BoolValue(dda.Features.OrchestratorExplorer.Enabled) {
		if !BoolValue(dda.Agent.Enabled) || dda.Agent.Process != nil {
			dda.Agent.Process.Enabled = NewBoolPointer(true)
			dso.Agent.Process = DefaultDatadogAgentSpecAgentProcess(&dda.Agent)
			dso.Agent.Process.Enabled = NewBoolPointer(true)
		}
	}
}

// DefaultDatadogAgentSpecAgent used to default an DatadogAgentSpecAgentSpec
// return the defaulted DatadogAgentSpecAgentSpec
func DefaultDatadogAgentSpecAgent(daemonsetAgents *DatadogAgentSpecAgentSpec) *DatadogAgentSpecAgentSpec {
	// If the Agent is not specified in the spec, disable it.
	if IsEqualStruct(*daemonsetAgents, DatadogAgentSpecAgentSpec{}) {
		daemonsetAgents.Enabled = NewBoolPointer(defaultAgentEnabled)

		if !BoolValue(daemonsetAgents.Enabled) {
			return daemonsetAgents
		}
	}

	daemonsetAgentsOverride := &DatadogAgentSpecAgentSpec{}
	if daemonsetAgents.Enabled == nil {
		daemonsetAgents.Enabled = NewBoolPointer(defaultAgentEnabled)
		daemonsetAgentsOverride.Enabled = daemonsetAgents.Enabled
	}

	if !BoolValue(daemonsetAgents.Enabled) {
		return daemonsetAgentsOverride
	}

	if daemonsetAgents.UseExtendedDaemonset == nil {
		daemonsetAgents.UseExtendedDaemonset = NewBoolPointer(false)
		daemonsetAgentsOverride.UseExtendedDaemonset = daemonsetAgents.UseExtendedDaemonset
	}

	if img := DefaultDatadogAgentSpecAgentImage(daemonsetAgents, defaultAgentImageName, defaultAgentImageTag); !IsEqualStruct(*img, ImageConfig{}) {
		daemonsetAgentsOverride.Image = img
	}

	if cfg := DefaultDatadogAgentSpecAgentConfig(daemonsetAgents); !IsEqualStruct(*cfg, NodeAgentSpec{}) {
		daemonsetAgentsOverride.NodeAgent = cfg
	}

	if rbac := DefaultDatadogAgentSpecRbacConfig(daemonsetAgents); !IsEqualStruct(*rbac, RbacConfig{}) {
		daemonsetAgentsOverride.Rbac = rbac
	}

	deployStrat := DefaultDatadogAgentSpecDatadogAgentStrategy(daemonsetAgents)
	if !IsEqualStruct(*deployStrat, DaemonSetDeploymentStrategy{}) {
		daemonsetAgentsOverride.DeploymentStrategy = deployStrat
	}

	if apm := DefaultDatadogAgentSpecAgentApm(daemonsetAgents); !IsEqualStruct(*apm, APMSpec{}) {
		daemonsetAgentsOverride.Apm = apm
	}

	if sysProb := DefaultDatadogAgentSpecAgentSystemProbe(daemonsetAgents); !IsEqualStruct(*sysProb, SystemProbeSpec{}) {
		daemonsetAgentsOverride.SystemProbe = sysProb
	}

	if sec := DefaultDatadogAgentSpecAgentSecurity(daemonsetAgents); !IsEqualStruct(*sec, SecuritySpec{}) {
		daemonsetAgentsOverride.Security = sec
	}

	if proc := DefaultDatadogAgentSpecAgentProcess(daemonsetAgents); !IsEqualStruct(*proc, ProcessSpec{}) {
		daemonsetAgentsOverride.Process = proc
	}

	if net := DefaultAgentNetworkPolicy(daemonsetAgents); !IsEqualStruct(*net, NetworkPolicySpec{}) {
		daemonsetAgentsOverride.NetworkPolicy = net
	}

	return daemonsetAgentsOverride
}

// DefaultDatadogAgentSpecAgentImage used to default a ImageConfig for the Agent, Cluster Agent and the Cluster Check Runner.
// Returns the defaulted ImageConfig.
func DefaultDatadogAgentSpecAgentImage(agent *DatadogAgentSpecAgentSpec, name, tag string) *ImageConfig {
	imgOverride := &ImageConfig{}
	if agent.Image == nil {
		agent.Image = &ImageConfig{}
	}

	if agent.Image.Name == "" {
		agent.Image.Name = name
		imgOverride.Name = agent.Image.Name
	}

	if agent.Image.Tag == "" {
		agent.Image.Tag = tag
		imgOverride.Tag = agent.Image.Tag
	}

	if agent.Image.PullPolicy == nil {
		agent.Image.PullPolicy = &defaultImagePullPolicy
		imgOverride.PullPolicy = agent.Image.PullPolicy
	}

	if agent.Image.PullSecrets == nil {
		agent.Image.PullSecrets = &[]corev1.LocalObjectReference{}
	}

	return imgOverride
}

// GetDefaultLivenessProbe creates a all defaulted LivenessProbe
func GetDefaultLivenessProbe() *corev1.Probe {
	livenessProbe := &corev1.Probe{
		InitialDelaySeconds: defaultLivenessProbeInitialDelaySeconds,
		PeriodSeconds:       defaultLivenessProbePeriodSeconds,
		TimeoutSeconds:      defaultLivenessProbeTimeoutSeconds,
		SuccessThreshold:    defaultLivenessProbeSuccessThreshold,
		FailureThreshold:    defaultLivenessProbeFailureThreshold,
	}
	livenessProbe.HTTPGet = &corev1.HTTPGetAction{
		Path: defaultLivenessProbeHTTPPath,
		Port: intstr.IntOrString{
			IntVal: defaultAgentHealthPort,
		},
	}
	return livenessProbe
}

// GetDefaultReadinessProbe creates a all defaulted ReadynessProbe
func GetDefaultReadinessProbe() *corev1.Probe {
	readinessProbe := &corev1.Probe{
		InitialDelaySeconds: defaultReadinessProbeInitialDelaySeconds,
		PeriodSeconds:       defaultReadinessProbePeriodSeconds,
		TimeoutSeconds:      defaultReadinessProbeTimeoutSeconds,
		SuccessThreshold:    defaultReadinessProbeSuccessThreshold,
		FailureThreshold:    defaultReadinessProbeFailureThreshold,
	}
	readinessProbe.HTTPGet = &corev1.HTTPGetAction{
		Path: defaultReadinessProbeHTTPPath,
		Port: intstr.IntOrString{
			IntVal: defaultAgentHealthPort,
		},
	}
	return readinessProbe
}

// DefaultDatadogAgentSpecAgentConfig used to default a NodeAgentConfig
// return the defaulted NodeAgentConfig
func DefaultDatadogAgentSpecAgentConfig(agents *DatadogAgentSpecAgentSpec) *NodeAgentSpec {
	configOverride := &NodeAgentSpec{}

	if agents.NodeAgent == nil {
		agents.NodeAgent = &NodeAgentSpec{}
	}

	if agents.NodeAgent.ContainerConfig.LogLevel == nil {
		agents.NodeAgent.ContainerConfig.LogLevel = NewStringPointer(defaultLogLevel)
		agents.NodeAgent.ContainerConfig.LogLevel = agents.NodeAgent.ContainerConfig.LogLevel
	}

	if agents.NodeAgent.CollectEvents == nil {
		agents.NodeAgent.CollectEvents = NewBoolPointer(defaultCollectEvents)
		configOverride.CollectEvents = agents.NodeAgent.CollectEvents
	}

	if agents.NodeAgent.LeaderElection == nil {
		agents.NodeAgent.LeaderElection = NewBoolPointer(defaultLeaderElection)
		configOverride.LeaderElection = agents.NodeAgent.LeaderElection
	}

	// Don't default Docker/CRI paths with Agent >= 7.27.0
	// Let Env AD do the work for us
	// Image is defaulted prior to this function.
	agentTag := strings.TrimSuffix(utils.GetTagFromImageName(agents.Image.Name), "-jmx")
	// Check against image tag + "-0"; otherwise prelease versions are not compared.
	// (See https://github.com/Masterminds/semver#working-with-prerelease-versions)
	if !(agentTag == "latest" || utils.IsAboveMinVersion(agentTag, "7.27.0-0") || utils.IsAboveMinVersion(agentTag, "6.27.0-0")) {
		if socketOverride := DefaultContainerSocket(agents.NodeAgent); !IsEqualStruct(socketOverride, CRISocketConfig{}) {
			configOverride.CriSocket = socketOverride
		}
	}

	if dsdOverride := DefaultConfigDogstatsd(agents.NodeAgent); !IsEqualStruct(dsdOverride, DogstatsdConfig{}) {
		configOverride.Dogstatsd = dsdOverride
	}

	if agents.NodeAgent.ContainerConfig.Resources == nil {
		agents.NodeAgent.ContainerConfig.Resources = &corev1.ResourceRequirements{}
	}

	if agents.NodeAgent.PodLabelsAsTags == nil {
		agents.NodeAgent.PodLabelsAsTags = map[string]string{}
	}

	if agents.NodeAgent.PodAnnotationsAsTags == nil {
		agents.NodeAgent.PodAnnotationsAsTags = map[string]string{}
	}

	if agents.NodeAgent.Tags == nil {
		agents.NodeAgent.Tags = []string{}
	}

	if agents.NodeAgent.ContainerConfig.LivenessProbe == nil {
		// TODO make liveness probe's fields more configurable
		agents.NodeAgent.ContainerConfig.LivenessProbe = GetDefaultLivenessProbe()
		configOverride.ContainerConfig.LivenessProbe = agents.NodeAgent.ContainerConfig.LivenessProbe
	}

	if agents.NodeAgent.ContainerConfig.ReadinessProbe == nil {
		// TODO make readiness probe's fields more configurable
		agents.NodeAgent.ContainerConfig.ReadinessProbe = GetDefaultReadinessProbe()
		configOverride.ContainerConfig.ReadinessProbe = agents.NodeAgent.ContainerConfig.ReadinessProbe
	}

	if agents.NodeAgent.ContainerConfig.HealthPort == nil {
		agents.NodeAgent.ContainerConfig.HealthPort = NewInt32Pointer(defaultAgentHealthPort)
		configOverride.ContainerConfig.HealthPort = agents.NodeAgent.ContainerConfig.HealthPort
	}

	return configOverride
}

// DefaultContainerSocket defaults the socket configuration for the Datadog Agent
func DefaultContainerSocket(config *NodeAgentSpec) *CRISocketConfig {
	if config.CriSocket == nil {
		config.CriSocket = &CRISocketConfig{
			DockerSocketPath: NewStringPointer(defaultDockerSocketPath),
		}
		return config.CriSocket
	}
	socketOverride := &CRISocketConfig{}
	if config.CriSocket.DockerSocketPath == nil {
		config.CriSocket.DockerSocketPath = NewStringPointer(defaultDockerSocketPath)
		socketOverride.DockerSocketPath = config.CriSocket.DockerSocketPath
	}
	return socketOverride
}

// DefaultConfigDogstatsd used to default Dogstatsd config in NodeAgentConfig
func DefaultConfigDogstatsd(config *NodeAgentSpec) *DogstatsdConfig {
	dsdOverride := &DogstatsdConfig{}
	if config.Dogstatsd == nil {
		config.Dogstatsd = &DogstatsdConfig{}
	}

	if config.Dogstatsd.DogstatsdOriginDetection == nil {
		config.Dogstatsd.DogstatsdOriginDetection = NewBoolPointer(defaultDogstatsdOriginDetection)
		dsdOverride.DogstatsdOriginDetection = config.Dogstatsd.DogstatsdOriginDetection
	}

	if uds := DefaultConfigDogstatsdUDS(config.Dogstatsd); !IsEqualStruct(uds, DSDUnixDomainSocketSpec{}) {
		dsdOverride.UnixDomainSocket = uds
	}

	return dsdOverride
}

// DefaultConfigDogstatsdUDS used to default DSDUnixDomainSocketSpec
// return the defaulted DSDUnixDomainSocketSpec
func DefaultConfigDogstatsdUDS(dsd *DogstatsdConfig) *DSDUnixDomainSocketSpec {
	if dsd.UnixDomainSocket == nil {
		dsd.UnixDomainSocket = &DSDUnixDomainSocketSpec{}
	}

	udsOverride := &DSDUnixDomainSocketSpec{}
	if dsd.UnixDomainSocket.Enabled == nil {
		dsd.UnixDomainSocket.Enabled = NewBoolPointer(defaultUseDogStatsDSocketVolume)
		udsOverride.Enabled = dsd.UnixDomainSocket.Enabled
	}

	if dsd.UnixDomainSocket.HostFilepath == nil {
		socketPath := path.Join(defaultHostDogstatsdSocketPath, defaultHostDogstatsdSocketName)
		dsd.UnixDomainSocket.HostFilepath = &socketPath
		udsOverride.HostFilepath = dsd.UnixDomainSocket.HostFilepath
	}

	return udsOverride
}

// DefaultRbacConfig defaults the RBAC section of the DatadogAgent
func DefaultRbacConfig(rbac *RbacConfig) *RbacConfig {
	rbacOverride := &RbacConfig{}
	if rbac == nil {
		rbac = &RbacConfig{}
	}

	if rbac.Create == nil {
		rbac.Create = NewBoolPointer(defaultRbacCreate)
		rbacOverride.Create = rbac.Create
	}

	return rbacOverride
}

// DefaultDatadogClusterCheckRunnerSpecRbacConfig used to default a RbacConfig of the Cluster Check Runner
func DefaultDatadogClusterCheckRunnerSpecRbacConfig(clc *DatadogAgentSpecClusterChecksRunnerSpec) *RbacConfig {
	if clc.Rbac == nil {
		// prevent passing an empty reference
		clc.Rbac = &RbacConfig{}
	}
	return DefaultRbacConfig(clc.Rbac)
}

// DefaultDatadogClusterAgentSpecRbacConfig used to default a RbacConfig of the Cluster Agent
func DefaultDatadogClusterAgentSpecRbacConfig(dca *DatadogAgentSpecClusterAgentSpec) *RbacConfig {
	if dca.Rbac == nil {
		// prevent passing an empty reference
		dca.Rbac = &RbacConfig{}
	}
	return DefaultRbacConfig(dca.Rbac)
}

// DefaultDatadogAgentSpecRbacConfig used to default a RbacConfig
// return the defaulted RbacConfig
func DefaultDatadogAgentSpecRbacConfig(agent *DatadogAgentSpecAgentSpec) *RbacConfig {
	if agent.Rbac == nil {
		// prevent passing an empty reference
		agent.Rbac = &RbacConfig{}
	}
	return DefaultRbacConfig(agent.Rbac)
}

// DefaultDatadogAgentSpecDatadogAgentStrategy used to default a DaemonSetDeploymentStrategy
// return the defaulted DaemonSetDeploymentStrategy
func DefaultDatadogAgentSpecDatadogAgentStrategy(agents *DatadogAgentSpecAgentSpec) *DaemonSetDeploymentStrategy {
	strategyOverride := &DaemonSetDeploymentStrategy{}
	if agents.DeploymentStrategy == nil {
		agents.DeploymentStrategy = &DaemonSetDeploymentStrategy{}
	}

	if agents.DeploymentStrategy.UpdateStrategyType == nil {
		updateStrategy := defaultUpdateStrategy
		agents.DeploymentStrategy.UpdateStrategyType = &updateStrategy
		strategyOverride.UpdateStrategyType = agents.DeploymentStrategy.UpdateStrategyType
	}

	if agents.DeploymentStrategy.RollingUpdate.MaxUnavailable == nil {
		agents.DeploymentStrategy.RollingUpdate.MaxUnavailable = &intstr.IntOrString{
			Type:   intstr.String,
			StrVal: defaultRollingUpdateMaxUnavailable,
		}
		strategyOverride.RollingUpdate.MaxUnavailable = agents.DeploymentStrategy.RollingUpdate.MaxUnavailable
	}

	if agents.DeploymentStrategy.RollingUpdate.MaxPodSchedulerFailure == nil {
		agents.DeploymentStrategy.RollingUpdate.MaxPodSchedulerFailure = &intstr.IntOrString{
			Type:   intstr.String,
			StrVal: defaultRollingUpdateMaxPodSchedulerFailure,
		}
		strategyOverride.RollingUpdate.MaxPodSchedulerFailure = agents.DeploymentStrategy.RollingUpdate.MaxPodSchedulerFailure
	}

	if agents.DeploymentStrategy.RollingUpdate.MaxParallelPodCreation == nil {
		agents.DeploymentStrategy.RollingUpdate.MaxParallelPodCreation = NewInt32Pointer(defaultRollingUpdateMaxParallelPodCreation)
		strategyOverride.RollingUpdate.MaxParallelPodCreation = agents.DeploymentStrategy.RollingUpdate.MaxParallelPodCreation
	}

	if agents.DeploymentStrategy.RollingUpdate.SlowStartIntervalDuration == nil {
		agents.DeploymentStrategy.RollingUpdate.SlowStartIntervalDuration = &metav1.Duration{
			Duration: defaultRollingUpdateSlowStartIntervalDuration,
		}
		strategyOverride.RollingUpdate.SlowStartIntervalDuration = agents.DeploymentStrategy.RollingUpdate.SlowStartIntervalDuration
	}

	if agents.DeploymentStrategy.RollingUpdate.SlowStartAdditiveIncrease == nil {
		agents.DeploymentStrategy.RollingUpdate.SlowStartAdditiveIncrease = &intstr.IntOrString{
			Type:   intstr.String,
			StrVal: defaultRollingUpdateSlowStartAdditiveIncrease,
		}
		strategyOverride.RollingUpdate.SlowStartAdditiveIncrease = agents.DeploymentStrategy.RollingUpdate.SlowStartAdditiveIncrease
	}

	if agents.DeploymentStrategy.Canary == nil {
		agents.DeploymentStrategy.Canary = edsdatadoghqv1alpha1.DefaultExtendedDaemonSetSpecStrategyCanary(&edsdatadoghqv1alpha1.ExtendedDaemonSetSpecStrategyCanary{})
		strategyOverride.Canary = agents.DeploymentStrategy.Canary
	}

	if agents.DeploymentStrategy.ReconcileFrequency == nil {
		agents.DeploymentStrategy.ReconcileFrequency = &metav1.Duration{
			Duration: defaultReconcileFrequency,
		}
		strategyOverride.ReconcileFrequency = agents.DeploymentStrategy.ReconcileFrequency
	}

	return strategyOverride
}

// DefaultDatadogAgentSpecAgentApm used to default an APMSpec
// return the defaulted APMSpec
func DefaultDatadogAgentSpecAgentApm(agents *DatadogAgentSpecAgentSpec) *APMSpec {
	if agents.Apm == nil {
		agents.Apm = &APMSpec{Enabled: NewBoolPointer(defaultApmEnabled)}
		return agents.Apm
	}

	apmOverride := &APMSpec{}
	if agents.Apm.Enabled == nil {
		agents.Apm.Enabled = NewBoolPointer(defaultApmEnabled)
		apmOverride.Enabled = agents.Apm.Enabled
	}

	if !BoolValue(agents.Apm.Enabled) {
		return apmOverride
	}

	if agents.Apm.HostPort == nil {
		agents.Apm.HostPort = NewInt32Pointer(defaultApmHostPort)
		apmOverride.HostPort = agents.Apm.HostPort
	}

	if agents.Apm.ContainerConfig.LivenessProbe == nil {
		agents.Apm.ContainerConfig.LivenessProbe = getDefaultAPMAgentLivenessProbe()
		apmOverride.ContainerConfig.LivenessProbe = agents.Apm.ContainerConfig.LivenessProbe
	}

	if udsOverride := DefaultDatadogAgentSpecAgentApmUDS(agents.Apm); !IsEqualStruct(udsOverride, APMUnixDomainSocketSpec{}) {
		apmOverride.UnixDomainSocket = udsOverride
	}

	return apmOverride
}

func getDefaultAPMAgentLivenessProbe() *corev1.Probe {
	livenessProbe := &corev1.Probe{
		InitialDelaySeconds: defaultLivenessProbeInitialDelaySeconds,
		PeriodSeconds:       defaultLivenessProbePeriodSeconds,
		TimeoutSeconds:      defaultLivenessProbeTimeoutSeconds,
	}
	livenessProbe.TCPSocket = &corev1.TCPSocketAction{
		Port: intstr.IntOrString{
			IntVal: defaultApmHostPort,
		},
	}
	return livenessProbe
}

// DefaultDatadogAgentSpecAgentApmUDS used to default APMUnixDomainSocketSpec
// rreturn the defaulted APMUnixDomainSocketSpec
func DefaultDatadogAgentSpecAgentApmUDS(apm *APMSpec) *APMUnixDomainSocketSpec {
	if apm.UnixDomainSocket == nil {
		apm.UnixDomainSocket = &APMUnixDomainSocketSpec{Enabled: NewBoolPointer(false)}
		return apm.UnixDomainSocket
	}

	udsOverride := &APMUnixDomainSocketSpec{}
	if apm.UnixDomainSocket.Enabled == nil {
		apm.UnixDomainSocket.Enabled = NewBoolPointer(false)
		udsOverride.Enabled = apm.UnixDomainSocket.Enabled
	}

	if !BoolValue(apm.UnixDomainSocket.Enabled) {
		return udsOverride
	}

	if apm.UnixDomainSocket.HostFilepath == nil {
		socketPath := path.Join(defaultHostApmSocketPath, defaultHostApmSocketName)
		apm.UnixDomainSocket.HostFilepath = &socketPath
		udsOverride.HostFilepath = apm.UnixDomainSocket.HostFilepath
	}

	return udsOverride
}

// DefaultDatadogAgentSpecAgentSystemProbe defaults the System Probe
// This method can be re-run as part of the FeatureOverride
func DefaultDatadogAgentSpecAgentSystemProbe(agents *DatadogAgentSpecAgentSpec) *SystemProbeSpec {
	if agents.SystemProbe == nil {
		agents.SystemProbe = &SystemProbeSpec{Enabled: NewBoolPointer(defaultSystemProbeEnabled)}
		return agents.SystemProbe
	}

	sysOverride := &SystemProbeSpec{}
	if agents.SystemProbe.Enabled == nil {
		agents.SystemProbe.Enabled = NewBoolPointer(defaultSystemProbeEnabled)
		sysOverride.Enabled = agents.SystemProbe.Enabled
	}

	if !BoolValue(agents.SystemProbe.Enabled) {
		return sysOverride
	}

	if agents.SystemProbe.EnableOOMKill == nil {
		agents.SystemProbe.EnableOOMKill = NewBoolPointer(defaultSystemProbeOOMKillEnabled)
		sysOverride.EnableOOMKill = agents.SystemProbe.EnableOOMKill
	}

	if agents.SystemProbe.EnableTCPQueueLength == nil {
		agents.SystemProbe.EnableTCPQueueLength = NewBoolPointer(defaultSystemProbeTCPQueueLengthEnabled)
		sysOverride.EnableTCPQueueLength = agents.SystemProbe.EnableTCPQueueLength
	}

	if agents.SystemProbe.BPFDebugEnabled == nil {
		agents.SystemProbe.BPFDebugEnabled = NewBoolPointer(defaultSystemProbeBPFDebugEnabled)
		sysOverride.BPFDebugEnabled = agents.SystemProbe.BPFDebugEnabled
	}

	if agents.SystemProbe.CollectDNSStats == nil {
		agents.SystemProbe.CollectDNSStats = NewBoolPointer(defaultSystemProbeCollectDNSStats)
		sysOverride.CollectDNSStats = agents.SystemProbe.CollectDNSStats
	}

	if agents.SystemProbe.ConntrackEnabled == nil {
		agents.SystemProbe.ConntrackEnabled = NewBoolPointer(defaultSystemProbeConntrackEnabled)
		sysOverride.ConntrackEnabled = agents.SystemProbe.ConntrackEnabled
	}

	if agents.SystemProbe.SecCompRootPath == "" {
		agents.SystemProbe.SecCompRootPath = defaultSystemProbeSecCompRootPath
		sysOverride.SecCompRootPath = agents.SystemProbe.SecCompRootPath
	}

	if agents.SystemProbe.AppArmorProfileName == "" {
		agents.SystemProbe.AppArmorProfileName = defaultAppArmorProfileName
		sysOverride.AppArmorProfileName = agents.SystemProbe.AppArmorProfileName
	}

	if agents.SystemProbe.SecCompProfileName == "" {
		agents.SystemProbe.SecCompProfileName = DefaultSeccompProfileName
		sysOverride.SecCompProfileName = agents.SystemProbe.SecCompProfileName
	}
	return sysOverride
}

// DefaultDatadogAgentSpecAgentSecurity defaults the Security Agent in the DatadogAgentSpec
func DefaultDatadogAgentSpecAgentSecurity(agents *DatadogAgentSpecAgentSpec) *SecuritySpec {
	secOverride := &SecuritySpec{}

	if agents.Security == nil {
		agents.Security = &SecuritySpec{}
	}

	if agents.Security.Compliance.Enabled == nil {
		agents.Security.Compliance.Enabled = NewBoolPointer(defaultSecurityComplianceEnabled)
		secOverride.Compliance.Enabled = agents.Security.Compliance.Enabled
	}

	if agents.Security.Runtime.Enabled == nil {
		agents.Security.Runtime.Enabled = NewBoolPointer(defaultSecurityRuntimeEnabled)
		secOverride.Runtime.Enabled = agents.Security.Runtime.Enabled
	}

	if agents.Security.Runtime.SyscallMonitor == nil {
		agents.Security.Runtime.SyscallMonitor = &SyscallMonitorSpec{}
		secOverride.Runtime.SyscallMonitor = agents.Security.Runtime.SyscallMonitor
	}

	if agents.Security.Runtime.SyscallMonitor.Enabled == nil {
		agents.Security.Runtime.SyscallMonitor.Enabled = NewBoolPointer(defaultSecuritySyscallMonitorEnabled)
		secOverride.Runtime.SyscallMonitor.Enabled = agents.Security.Runtime.SyscallMonitor.Enabled
	}

	return secOverride
}

// DefaultDatadogFeatureLogCollection used to default an LogCollectionConfig
// return the defaulted LogCollectionConfig
func DefaultDatadogFeatureLogCollection(ft *DatadogFeatures) *LogCollectionConfig {
	if ft.LogCollection == nil {
		ft.LogCollection = &LogCollectionConfig{Enabled: NewBoolPointer(defaultLogEnabled)}
		return ft.LogCollection
	}

	if ft.LogCollection.Enabled == nil {
		ft.LogCollection.Enabled = NewBoolPointer(defaultLogEnabled)
	}

	logOverride := &LogCollectionConfig{Enabled: ft.LogCollection.Enabled}

	if !BoolValue(ft.LogCollection.Enabled) {
		return logOverride
	}

	if ft.LogCollection.LogsConfigContainerCollectAll == nil {
		ft.LogCollection.LogsConfigContainerCollectAll = NewBoolPointer(defaultLogsConfigContainerCollectAll)
		logOverride.LogsConfigContainerCollectAll = ft.LogCollection.LogsConfigContainerCollectAll
	}

	if ft.LogCollection.ContainerCollectUsingFiles == nil {
		ft.LogCollection.ContainerCollectUsingFiles = NewBoolPointer(defaultLogsContainerCollectUsingFiles)
		logOverride.ContainerCollectUsingFiles = ft.LogCollection.ContainerCollectUsingFiles
	}

	if ft.LogCollection.ContainerLogsPath == nil {
		ft.LogCollection.ContainerLogsPath = NewStringPointer(defaultContainerLogsPath)
		logOverride.ContainerLogsPath = ft.LogCollection.ContainerLogsPath
	}

	if ft.LogCollection.PodLogsPath == nil {
		ft.LogCollection.PodLogsPath = NewStringPointer(defaultPodLogsPath)
		logOverride.PodLogsPath = ft.LogCollection.PodLogsPath
	}

	if ft.LogCollection.ContainerSymlinksPath == nil {
		ft.LogCollection.ContainerSymlinksPath = NewStringPointer(defaultContainerSymlinksPath)
		logOverride.ContainerSymlinksPath = ft.LogCollection.ContainerSymlinksPath
	}

	if ft.LogCollection.TempStoragePath == nil {
		ft.LogCollection.TempStoragePath = NewStringPointer(defaultLogsTempStoragePath)
		logOverride.TempStoragePath = ft.LogCollection.TempStoragePath
	}

	if ft.LogCollection.OpenFilesLimit == nil {
		ft.LogCollection.OpenFilesLimit = NewInt32Pointer(defaultLogsOpenFilesLimit)
		logOverride.OpenFilesLimit = ft.LogCollection.OpenFilesLimit
	}

	return logOverride
}

// DefaultDatadogAgentSpecAgentProcess used to default an ProcessSpec
// return the defaulted ProcessSpec
func DefaultDatadogAgentSpecAgentProcess(agents *DatadogAgentSpecAgentSpec) *ProcessSpec {
	if agents.Process == nil {
		agents.Process = &ProcessSpec{Enabled: NewBoolPointer(defaultProcessEnabled)}
		return agents.Process
	}

	processOverride := &ProcessSpec{}

	if agents.Process.Enabled == nil {
		agents.Process.Enabled = NewBoolPointer(defaultProcessEnabled)
		processOverride.Enabled = agents.Process.Enabled
	}

	if !BoolValue(agents.Process.Enabled) {
		return processOverride
	}

	if agents.Process.ProcessCollectionEnabled == nil {
		agents.Process.ProcessCollectionEnabled = NewBoolPointer(defaultProcessCollectionEnabled)
		processOverride.ProcessCollectionEnabled = agents.Process.ProcessCollectionEnabled
	}

	return processOverride
}

func clusterChecksRunnerEnabled(dda *DatadogAgent) bool {
	if dda.Spec.ClusterChecksRunner.Enabled != nil {
		return *dda.Spec.ClusterChecksRunner.Enabled
	}

	return false
}

// DefaultFeatures used to initialized the Features' default values if necessary
func DefaultFeatures(dda *DatadogAgent) *DatadogFeatures {
	ft := &dda.Spec.Features
	featureOverride := &DatadogFeatures{}

	clusterCheckEnabled := clusterChecksRunnerEnabled(dda)

	if orch := DefaultDatadogFeatureOrchestratorExplorer(ft, clusterCheckEnabled); !IsEqualStruct(*orch, OrchestratorExplorerConfig{}) {
		featureOverride.OrchestratorExplorer = orch
	}

	if ksm := DefaultDatadogFeatureKubeStateMetricsCore(ft, clusterCheckEnabled); !IsEqualStruct(*ksm, KubeStateMetricsCore{}) {
		featureOverride.KubeStateMetricsCore = ksm
	}

	if promScrape := DefaultDatadogFeaturePrometheusScrape(ft); !IsEqualStruct(*promScrape, PrometheusScrapeConfig{}) {
		featureOverride.PrometheusScrape = promScrape
	}

	if logColl := DefaultDatadogFeatureLogCollection(ft); !IsEqualStruct(*logColl, LogCollectionConfig{}) {
		featureOverride.LogCollection = logColl
	}

	if net := DefaultDatadogFeatureNetworkMonitoring(ft); !IsEqualStruct(*net, NetworkMonitoringConfig{}) {
		featureOverride.NetworkMonitoring = net
	}

	return featureOverride
}

// DefaultDatadogFeatureOrchestratorExplorer used to default an OrchestratorExplorerConfig
// return the defaulted OrchestratorExplorerConfig
func DefaultDatadogFeatureOrchestratorExplorer(ft *DatadogFeatures, withClusterChecksRunner bool) *OrchestratorExplorerConfig {
	if ft.OrchestratorExplorer == nil {
		ft.OrchestratorExplorer = &OrchestratorExplorerConfig{}
	}

	return defaultEnabledDatadogFeatureOrchestratorExplorer(ft.OrchestratorExplorer, withClusterChecksRunner)
}

func defaultEnabledDatadogFeatureOrchestratorExplorer(config *OrchestratorExplorerConfig, withClusterChecksRunner bool) *OrchestratorExplorerConfig {
	explorerConfigOverride := &OrchestratorExplorerConfig{}

	if config.Enabled == nil {
		config.Enabled = NewBoolPointer(defaultOrchestratorExplorerEnabled)
		explorerConfigOverride.Enabled = config.Enabled
	}
	if BoolValue(config.Enabled) {
		if config.ClusterCheck == nil {
			config.ClusterCheck = NewBoolPointer(withClusterChecksRunner)
			explorerConfigOverride.ClusterCheck = config.ClusterCheck
		}

		if config.Scrubbing == nil {
			config.Scrubbing = &Scrubbing{}
			explorerConfigOverride.Scrubbing = config.Scrubbing
		}

		if config.Scrubbing.Containers == nil {
			config.Scrubbing.Containers = NewBoolPointer(defaultOrchestratorExplorerContainerScrubbingEnabled)
			explorerConfigOverride.Scrubbing.Containers = config.Scrubbing.Containers
		}
	} else {
		explorerConfigOverride.Enabled = NewBoolPointer(false)
	}
	return explorerConfigOverride
}

// DefaultDatadogFeatureKubeStateMetricsCore used to default the Kubernetes State Metrics core check
// Disabled by default with no overridden configuration.
func DefaultDatadogFeatureKubeStateMetricsCore(ft *DatadogFeatures, withClusterChecksRunner bool) *KubeStateMetricsCore {
	if ft.KubeStateMetricsCore == nil {
		ft.KubeStateMetricsCore = &KubeStateMetricsCore{
			Enabled:      NewBoolPointer(defaultKubeStateMetricsCoreEnabled),
			ClusterCheck: NewBoolPointer(withClusterChecksRunner),
		}
		return ft.KubeStateMetricsCore
	}

	if ft.KubeStateMetricsCore.Enabled == nil {
		ft.KubeStateMetricsCore.Enabled = NewBoolPointer(defaultKubeStateMetricsCoreEnabled)
	}

	if ft.KubeStateMetricsCore.ClusterCheck == nil {
		ft.KubeStateMetricsCore.ClusterCheck = NewBoolPointer(withClusterChecksRunner)
	}

	ksmCoreOverride := &KubeStateMetricsCore{Enabled: ft.KubeStateMetricsCore.Enabled, ClusterCheck: ft.KubeStateMetricsCore.ClusterCheck}
	return ksmCoreOverride
}

// DefaultDatadogFeaturePrometheusScrape used to default the Prometheus Scrape config
func DefaultDatadogFeaturePrometheusScrape(ft *DatadogFeatures) *PrometheusScrapeConfig {
	if ft.PrometheusScrape == nil {
		ft.PrometheusScrape = &PrometheusScrapeConfig{Enabled: NewBoolPointer(defaultPrometheusScrapeEnabled)}
	}

	if ft.PrometheusScrape.Enabled == nil {
		ft.PrometheusScrape.Enabled = NewBoolPointer(defaultPrometheusScrapeEnabled)
	}

	promOverride := &PrometheusScrapeConfig{Enabled: ft.PrometheusScrape.Enabled}

	if !BoolValue(ft.PrometheusScrape.Enabled) {
		return promOverride
	}

	if ft.PrometheusScrape.ServiceEndpoints == nil {
		ft.PrometheusScrape.ServiceEndpoints = NewBoolPointer(defaultPrometheusScrapeServiceEndpoints)
		promOverride.ServiceEndpoints = ft.PrometheusScrape.ServiceEndpoints
	}

	return promOverride
}

// DefaultDatadogFeatureNetworkMonitoring used to default the NetworkMonitoring config
func DefaultDatadogFeatureNetworkMonitoring(ft *DatadogFeatures) *NetworkMonitoringConfig {
	if ft.NetworkMonitoring == nil {
		ft.NetworkMonitoring = &NetworkMonitoringConfig{Enabled: NewBoolPointer(false)}

		if !BoolValue(ft.NetworkMonitoring.Enabled) {
			return ft.NetworkMonitoring
		}
	}

	if ft.NetworkMonitoring.Enabled == nil {
		ft.NetworkMonitoring.Enabled = NewBoolPointer(false)
	}

	netOverride := &NetworkMonitoringConfig{Enabled: ft.NetworkMonitoring.Enabled}

	return netOverride
}

// DefaultDatadogAgentSpecClusterAgent used to default an DatadogAgentSpecClusterAgentSpec
// Mutate the internal DatadogAgentSpecClusterAgent throughout the method
// return the defaulted DatadogAgentSpecClusterAgentSpec to update the status
func DefaultDatadogAgentSpecClusterAgent(clusterAgent *DatadogAgentSpecClusterAgentSpec) *DatadogAgentSpecClusterAgentSpec {
	if IsEqualStruct(*clusterAgent, DatadogAgentSpecClusterAgentSpec{}) {
		clusterAgent.Enabled = NewBoolPointer(defaultClusterAgentEnabled)

		if !BoolValue(clusterAgent.Enabled) {
			return clusterAgent
		}
	}

	clusterAgentOverride := &DatadogAgentSpecClusterAgentSpec{}

	if clusterAgent.Enabled == nil {
		// Cluster Agent is enabled by default unless undeclared then it is set to false.
		clusterAgent.Enabled = NewBoolPointer(defaultClusterAgentEnabled)
		clusterAgentOverride.Enabled = clusterAgent.Enabled
	}

	if !BoolValue(clusterAgent.Enabled) {
		return clusterAgentOverride
	}

	if clusterAgent.Image == nil {
		clusterAgent.Image = &ImageConfig{}
	}
	if img := DefaultDatadogClusterAgentImage(clusterAgent, defaultClusterAgentImageName, defaultClusterAgentImageTag); !IsEqualStruct(*img, ImageConfig{}) {
		clusterAgentOverride.Image = img
	}

	if cfg := DefaultDatadogAgentSpecClusterAgentConfig(clusterAgent); !IsEqualStruct(cfg, ClusterAgentConfig{}) {
		clusterAgentOverride.Config = cfg
	}

	if rbac := DefaultDatadogClusterAgentSpecRbacConfig(clusterAgent); !IsEqualStruct(rbac, RbacConfig{}) {
		clusterAgentOverride.Rbac = rbac
	}

	if net := DefaultClusterAgentNetworkPolicy(clusterAgent); !IsEqualStruct(net, NetworkPolicySpec{}) {
		clusterAgentOverride.NetworkPolicy = net
	}

	return clusterAgentOverride
}

// DefaultDatadogAgentSpecClusterAgentConfig used to default an ClusterAgentConfig
// return the defaulted ClusterAgentConfig
func DefaultDatadogAgentSpecClusterAgentConfig(dca *DatadogAgentSpecClusterAgentSpec) *ClusterAgentConfig {
	configOverride := &ClusterAgentConfig{}

	if dca.CustomConfig == nil {
		dca.Config = &ClusterAgentConfig{}
	}

	if dca.Config.ContainerConfig.LogLevel == nil {
		dca.Config.ContainerConfig.LogLevel = NewStringPointer(defaultLogLevel)
		configOverride.ContainerConfig.LogLevel = dca.Config.ContainerConfig.LogLevel
	}

	if extMetricsOverride := DefaultExternalMetrics(dca.Config.Features); !IsEqualStruct(extMetricsOverride, ExternalMetricsConfig{}) {
		configOverride.Features.ExternalMetrics = extMetricsOverride
	}

	if dca.Config.Features.ClusterChecksEnabled == nil {
		dca.Config.Features.ClusterChecksEnabled = NewBoolPointer(defaultClusterChecksEnabled)
		configOverride.Features.ClusterChecksEnabled = dca.Config.Features.ClusterChecksEnabled
	}

	if dca.Config.Features.CollectEvents == nil {
		dca.Config.Features.CollectEvents = NewBoolPointer(defaultCollectEvents)
		configOverride.Features.CollectEvents = dca.Config.Features.CollectEvents
	}
	if admCtrlOverride := DefaultAdmissionController(dca.Config.Features); !IsEqualStruct(admCtrlOverride, AdmissionControllerConfig{}) {
		configOverride.Features.AdmissionController = admCtrlOverride
	}

	if dca.Config.ContainerConfig.Resources == nil {
		dca.Config.ContainerConfig.Resources = &corev1.ResourceRequirements{}
	}

	if dca.Config.ContainerConfig.HealthPort == nil {
		dca.Config.ContainerConfig.HealthPort = NewInt32Pointer(defaultAgentHealthPort)
		configOverride.ContainerConfig.HealthPort = dca.Config.ContainerConfig.HealthPort
	}

	return configOverride
}

// DefaultExternalMetrics defaults the External Metrics Server's config in the Cluster Agent's config
func DefaultExternalMetrics(conf *FeaturesConfigClusterAgent) *ExternalMetricsConfig {
	if conf.ExternalMetrics == nil {
		conf.ExternalMetrics = &ExternalMetricsConfig{Enabled: NewBoolPointer(defaultExternalMetricsEnabled)}

		if !BoolValue(conf.ExternalMetrics.Enabled) {
			return conf.ExternalMetrics
		}
	}

	extMetricsOverride := &ExternalMetricsConfig{}
	if conf.ExternalMetrics.Enabled == nil {
		// default to `true` because in that case we know that other parameters are
		// present in the `conf.ExternalMetrics` struct.
		conf.ExternalMetrics.Enabled = NewBoolPointer(true)
		extMetricsOverride.Enabled = conf.ExternalMetrics.Enabled
	}

	if conf.ExternalMetrics.Port == nil && BoolValue(conf.ExternalMetrics.Enabled) {
		conf.ExternalMetrics.Port = NewInt32Pointer(defaultMetricsProviderPort)
		extMetricsOverride.Port = conf.ExternalMetrics.Port
	}
	return extMetricsOverride
}

// DefaultAdmissionController defaults the Admission Controller's config in the Cluster Agent's config
func DefaultAdmissionController(conf *FeaturesConfigClusterAgent) *AdmissionControllerConfig {
	if conf.AdmissionController == nil {
		conf.AdmissionController = &AdmissionControllerConfig{Enabled: NewBoolPointer(defaultAdmissionControllerEnabled)}

		if !BoolValue(conf.AdmissionController.Enabled) {
			return conf.AdmissionController
		}
	}
	admCtrlOverride := &AdmissionControllerConfig{}

	if conf.AdmissionController.Enabled == nil {
		conf.AdmissionController.Enabled = NewBoolPointer(defaultAdmissionControllerEnabled)
		admCtrlOverride.Enabled = conf.AdmissionController.Enabled
	}

	if conf.AdmissionController.MutateUnlabelled == nil {
		conf.AdmissionController.MutateUnlabelled = NewBoolPointer(defaultMutateUnlabelled)
		admCtrlOverride.MutateUnlabelled = conf.AdmissionController.MutateUnlabelled
	}

	if conf.AdmissionController.ServiceName == nil {
		conf.AdmissionController.ServiceName = NewStringPointer(DefaultAdmissionServiceName)
		admCtrlOverride.ServiceName = conf.AdmissionController.ServiceName
	}
	return admCtrlOverride
}

// DefaultDatadogClusterAgentImage used to default a ImageConfig for the Agent, Cluster Agent and the Cluster Check Runner.
// Returns the defaulted ImageConfig.
func DefaultDatadogClusterAgentImage(dca *DatadogAgentSpecClusterAgentSpec, name, tag string) *ImageConfig {
	imgOverride := &ImageConfig{}
	if dca.Image == nil {
		dca.Image = &ImageConfig{}
	}

	if dca.Image.Name == "" {
		dca.Image.Name = name
		imgOverride.Name = dca.Image.Name
	}

	if dca.Image.Tag == "" {
		dca.Image.Tag = tag
		imgOverride.Tag = dca.Image.Tag
	}

	if dca.Image.PullPolicy == nil {
		dca.Image.PullPolicy = &defaultImagePullPolicy
		imgOverride.PullPolicy = dca.Image.PullPolicy
	}

	if dca.Image.PullSecrets == nil {
		dca.Image.PullSecrets = &[]corev1.LocalObjectReference{}
	}

	return imgOverride
}

// DefaultDatadogAgentSpecClusterChecksRunner used to default an DatadogAgentSpecClusterChecksRunnerSpec
// return the defaulted DatadogAgentSpecClusterChecksRunnerSpec
func DefaultDatadogAgentSpecClusterChecksRunner(clusterChecksRunner *DatadogAgentSpecClusterChecksRunnerSpec) *DatadogAgentSpecClusterChecksRunnerSpec {
	if IsEqualStruct(clusterChecksRunner, DatadogAgentSpecClusterChecksRunnerSpec{}) {
		clusterChecksRunner.Enabled = NewBoolPointer(defaultClusterChecksRunnerEnabled)

		if !BoolValue(clusterChecksRunner.Enabled) {
			return clusterChecksRunner
		}
	}

	clcOverride := &DatadogAgentSpecClusterChecksRunnerSpec{}
	if clusterChecksRunner.Enabled == nil {
		// Default to `true` because we are in the case it means other parameters
		// are present in the struct.
		clusterChecksRunner.Enabled = NewBoolPointer(true)
		clcOverride.Enabled = clusterChecksRunner.Enabled
	}

	if img := DefaultDatadogAgentSpecClusterChecksRunnerImage(clusterChecksRunner, defaultAgentImageName, defaultAgentImageTag); !IsEqualStruct(img, ImageConfig{}) {
		clcOverride.Image = img
	}

	if cfg := DefaultDatadogAgentSpecClusterChecksRunnerConfig(clusterChecksRunner); !IsEqualStruct(cfg, ClusterChecksRunnerConfig{}) {
		clcOverride.ContainerConfig = cfg
	}

	if rbac := DefaultDatadogClusterCheckRunnerSpecRbacConfig(clusterChecksRunner); !IsEqualStruct(rbac, RbacConfig{}) {
		clcOverride.Rbac = rbac
	}

	if net := DefaultClusterCheckRunnerNetworkPolicy(clusterChecksRunner); !IsEqualStruct(net, NetworkPolicySpec{}) {
		clcOverride.NetworkPolicy = net
	}

	return clcOverride
}

// DefaultDatadogAgentSpecClusterChecksRunnerImage used to default a ImageConfig for the Agent, Cluster Agent and the Cluster Check Runner.
// Returns the defaulted ImageConfig.
func DefaultDatadogAgentSpecClusterChecksRunnerImage(clc *DatadogAgentSpecClusterChecksRunnerSpec, name, tag string) *ImageConfig {
	imgOverride := &ImageConfig{}
	if clc.Image == nil {
		clc.Image = &ImageConfig{}
	}

	if clc.Image.Name == "" {
		clc.Image.Name = name
		imgOverride.Name = clc.Image.Name
	}

	if clc.Image.Tag == "" {
		clc.Image.Tag = tag
		imgOverride.Tag = clc.Image.Tag
	}

	if clc.Image.PullPolicy == nil {
		clc.Image.PullPolicy = &defaultImagePullPolicy
		imgOverride.PullPolicy = clc.Image.PullPolicy
	}

	if clc.Image.PullSecrets == nil {
		clc.Image.PullSecrets = &[]corev1.LocalObjectReference{}
	}

	return imgOverride
}

// DefaultDatadogAgentSpecClusterChecksRunnerConfig used to default an ClusterChecksRunnerConfig
// return the defaulted ClusterChecksRunnerConfig
func DefaultDatadogAgentSpecClusterChecksRunnerConfig(clc *DatadogAgentSpecClusterChecksRunnerSpec) *DatadogAgentGenericContainerConfig {
	configOverride := &DatadogAgentGenericContainerConfig{}

	if clc.ContainerConfig == nil {
		clc.ContainerConfig = &DatadogAgentGenericContainerConfig{}
	}

	if clc.ContainerConfig.LogLevel == nil {
		clc.ContainerConfig.LogLevel = NewStringPointer(defaultLogLevel)
		configOverride.LogLevel = clc.ContainerConfig.LogLevel
	}

	if clc.ContainerConfig.LivenessProbe == nil {
		// TODO make liveness probe's fields more configurable
		clc.ContainerConfig.LivenessProbe = GetDefaultLivenessProbe()
		configOverride.LivenessProbe = clc.ContainerConfig.LivenessProbe
	}

	if clc.ContainerConfig.ReadinessProbe == nil {
		// TODO make readiness probe's fields more configurable
		clc.ContainerConfig.ReadinessProbe = GetDefaultReadinessProbe()
		configOverride.ReadinessProbe = clc.ContainerConfig.ReadinessProbe
	}
	if clc.ContainerConfig.HealthPort == nil {
		clc.ContainerConfig.HealthPort = NewInt32Pointer(defaultAgentHealthPort)
		configOverride.HealthPort = clc.ContainerConfig.HealthPort
	}

	if clc.ContainerConfig.Resources == nil {
		clc.ContainerConfig.Resources = &corev1.ResourceRequirements{}
	}
	return configOverride
}

// DefaultNetworkPolicy is used to default NetworkPolicy. Returns the defaulted
// NetworkPolicySpec
func DefaultNetworkPolicy(policy *NetworkPolicySpec) *NetworkPolicySpec {
	policyOverride := &NetworkPolicySpec{}
	if policy == nil {
		policy = &NetworkPolicySpec{}
	}

	if policy.Create == nil {
		policy.Create = NewBoolPointer(false)
		policyOverride.Create = policy.Create
	}

	return policyOverride
}

// DefaultAgentNetworkPolicy defaults the Network Policy for the Datadog Agent
func DefaultAgentNetworkPolicy(agent *DatadogAgentSpecAgentSpec) *NetworkPolicySpec {
	if agent.NetworkPolicy == nil {
		agent.NetworkPolicy = &NetworkPolicySpec{}
	}
	return DefaultNetworkPolicy(agent.NetworkPolicy)
}

// DefaultClusterAgentNetworkPolicy defaults the Network Policy for the Datadog Cluster Agent
func DefaultClusterAgentNetworkPolicy(dca *DatadogAgentSpecClusterAgentSpec) *NetworkPolicySpec {
	if dca.NetworkPolicy == nil {
		dca.NetworkPolicy = &NetworkPolicySpec{}
	}
	return DefaultNetworkPolicy(dca.NetworkPolicy)
}

// DefaultClusterCheckRunnerNetworkPolicy defaults the Network Policy for the Cluster Check Runner
func DefaultClusterCheckRunnerNetworkPolicy(clc *DatadogAgentSpecClusterChecksRunnerSpec) *NetworkPolicySpec {
	if clc.NetworkPolicy == nil {
		clc.NetworkPolicy = &NetworkPolicySpec{}
	}
	return DefaultNetworkPolicy(clc.NetworkPolicy)
}
