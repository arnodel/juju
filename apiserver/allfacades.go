// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package apiserver

import (
	"reflect"

	"github.com/juju/errors"
	"github.com/juju/names/v4"

	apiservererrors "github.com/juju/juju/apiserver/errors"
	"github.com/juju/juju/apiserver/facade"
	"github.com/juju/juju/apiserver/facades/agent/agent"
	"github.com/juju/juju/apiserver/facades/agent/caasadmission"
	"github.com/juju/juju/apiserver/facades/agent/caasagent"
	"github.com/juju/juju/apiserver/facades/agent/caasapplication"
	"github.com/juju/juju/apiserver/facades/agent/caasoperator"
	"github.com/juju/juju/apiserver/facades/agent/credentialvalidator"
	"github.com/juju/juju/apiserver/facades/agent/deployer"
	"github.com/juju/juju/apiserver/facades/agent/diskmanager"
	"github.com/juju/juju/apiserver/facades/agent/fanconfigurer"
	"github.com/juju/juju/apiserver/facades/agent/hostkeyreporter"
	"github.com/juju/juju/apiserver/facades/agent/instancemutater"
	"github.com/juju/juju/apiserver/facades/agent/keyupdater"
	"github.com/juju/juju/apiserver/facades/agent/leadership"
	loggerapi "github.com/juju/juju/apiserver/facades/agent/logger"
	"github.com/juju/juju/apiserver/facades/agent/machine"
	"github.com/juju/juju/apiserver/facades/agent/machineactions"
	"github.com/juju/juju/apiserver/facades/agent/meterstatus"
	"github.com/juju/juju/apiserver/facades/agent/metricsadder"
	"github.com/juju/juju/apiserver/facades/agent/migrationflag"
	"github.com/juju/juju/apiserver/facades/agent/migrationminion"
	"github.com/juju/juju/apiserver/facades/agent/payloadshookcontext"
	"github.com/juju/juju/apiserver/facades/agent/provisioner"
	"github.com/juju/juju/apiserver/facades/agent/proxyupdater"
	"github.com/juju/juju/apiserver/facades/agent/reboot"
	"github.com/juju/juju/apiserver/facades/agent/resourceshookcontext"
	"github.com/juju/juju/apiserver/facades/agent/retrystrategy"
	"github.com/juju/juju/apiserver/facades/agent/secretsmanager"
	"github.com/juju/juju/apiserver/facades/agent/storageprovisioner"
	"github.com/juju/juju/apiserver/facades/agent/unitassigner"
	"github.com/juju/juju/apiserver/facades/agent/uniter"
	"github.com/juju/juju/apiserver/facades/agent/upgrader"
	"github.com/juju/juju/apiserver/facades/agent/upgradeseries"
	"github.com/juju/juju/apiserver/facades/agent/upgradesteps"
	"github.com/juju/juju/apiserver/facades/client/action"
	"github.com/juju/juju/apiserver/facades/client/annotations" // ModelUser Write
	"github.com/juju/juju/apiserver/facades/client/application" // ModelUser Write
	"github.com/juju/juju/apiserver/facades/client/applicationoffers"
	"github.com/juju/juju/apiserver/facades/client/backups" // ModelUser Write
	"github.com/juju/juju/apiserver/facades/client/block"   // ModelUser Write
	"github.com/juju/juju/apiserver/facades/client/bundle"
	"github.com/juju/juju/apiserver/facades/client/charms"     // ModelUser Write
	"github.com/juju/juju/apiserver/facades/client/client"     // ModelUser Write
	"github.com/juju/juju/apiserver/facades/client/cloud"      // ModelUser Read
	"github.com/juju/juju/apiserver/facades/client/controller" // ModelUser Admin (although some methods check for read only)
	"github.com/juju/juju/apiserver/facades/client/credentialmanager"
	"github.com/juju/juju/apiserver/facades/client/firewallrules"
	"github.com/juju/juju/apiserver/facades/client/highavailability" // ModelUser Write
	"github.com/juju/juju/apiserver/facades/client/imagemanager"     // ModelUser Write
	"github.com/juju/juju/apiserver/facades/client/imagemetadatamanager"
	"github.com/juju/juju/apiserver/facades/client/keymanager"     // ModelUser Write
	"github.com/juju/juju/apiserver/facades/client/machinemanager" // ModelUser Write
	"github.com/juju/juju/apiserver/facades/client/metricsdebug"   // ModelUser Write
	"github.com/juju/juju/apiserver/facades/client/modelconfig"    // ModelUser Write
	"github.com/juju/juju/apiserver/facades/client/modelgeneration"
	"github.com/juju/juju/apiserver/facades/client/modelmanager" // ModelUser Write
	"github.com/juju/juju/apiserver/facades/client/payloads"
	"github.com/juju/juju/apiserver/facades/client/resources"
	"github.com/juju/juju/apiserver/facades/client/secrets"
	"github.com/juju/juju/apiserver/facades/client/spaces"    // ModelUser Write
	"github.com/juju/juju/apiserver/facades/client/sshclient" // ModelUser Write
	"github.com/juju/juju/apiserver/facades/client/storage"
	"github.com/juju/juju/apiserver/facades/client/subnets"
	"github.com/juju/juju/apiserver/facades/client/usermanager"
	"github.com/juju/juju/apiserver/facades/controller/actionpruner"
	"github.com/juju/juju/apiserver/facades/controller/agenttools"
	"github.com/juju/juju/apiserver/facades/controller/applicationscaler"
	"github.com/juju/juju/apiserver/facades/controller/caasapplicationprovisioner"
	"github.com/juju/juju/apiserver/facades/controller/caasfirewaller"
	"github.com/juju/juju/apiserver/facades/controller/caasmodelconfigmanager"
	"github.com/juju/juju/apiserver/facades/controller/caasmodeloperator"
	"github.com/juju/juju/apiserver/facades/controller/caasoperatorprovisioner"
	"github.com/juju/juju/apiserver/facades/controller/caasoperatorupgrader"
	"github.com/juju/juju/apiserver/facades/controller/caasunitprovisioner"
	"github.com/juju/juju/apiserver/facades/controller/charmdownloader"
	"github.com/juju/juju/apiserver/facades/controller/charmrevisionupdater"
	"github.com/juju/juju/apiserver/facades/controller/cleaner"
	"github.com/juju/juju/apiserver/facades/controller/crosscontroller"
	"github.com/juju/juju/apiserver/facades/controller/crossmodelrelations"
	"github.com/juju/juju/apiserver/facades/controller/externalcontrollerupdater"
	"github.com/juju/juju/apiserver/facades/controller/firewaller"
	"github.com/juju/juju/apiserver/facades/controller/imagemetadata"
	"github.com/juju/juju/apiserver/facades/controller/instancepoller"
	"github.com/juju/juju/apiserver/facades/controller/lifeflag"
	"github.com/juju/juju/apiserver/facades/controller/logfwd"
	"github.com/juju/juju/apiserver/facades/controller/machineundertaker"
	"github.com/juju/juju/apiserver/facades/controller/metricsmanager"
	"github.com/juju/juju/apiserver/facades/controller/migrationmaster"
	"github.com/juju/juju/apiserver/facades/controller/migrationtarget"
	"github.com/juju/juju/apiserver/facades/controller/modelupgrader"
	"github.com/juju/juju/apiserver/facades/controller/raftlease"
	"github.com/juju/juju/apiserver/facades/controller/remoterelations"
	"github.com/juju/juju/apiserver/facades/controller/resumer"
	"github.com/juju/juju/apiserver/facades/controller/singular"
	"github.com/juju/juju/apiserver/facades/controller/statushistory"
	"github.com/juju/juju/apiserver/facades/controller/undertaker"
	"github.com/juju/juju/state"
)

// AllFacades returns a registry containing all known API facades.
//
// This will panic if facade registration fails, but there is a unit
// test to guard against that.
func AllFacades() *facade.Registry {
	registry := new(facade.Registry)

	reg := func(name string, version int, newFunc interface{}) {
		err := registry.RegisterStandard(name, version, newFunc)
		if err != nil {
			panic(err)
		}
	}

	regRaw := func(name string, version int, factory facade.Factory, facadeType reflect.Type) {
		err := registry.Register(name, version, factory, facadeType)
		if err != nil {
			panic(err)
		}
	}

	regHookContext := func(name string, version int, newHookContextFacade hookContextFacadeFn, facadeType reflect.Type) {
		err := regHookContextFacade(registry, name, version, newHookContextFacade, facadeType)
		if err != nil {
			panic(err)
		}
	}

	reg("Action", 7, action.NewActionAPIV7)
	reg("ActionPruner", 1, actionpruner.NewAPI)
	reg("Agent", 3, agent.NewAgentAPIV3)
	reg("AgentTools", 1, agenttools.NewFacade)
	reg("Annotations", 2, annotations.NewAPI)

	reg("Application", 13, application.NewFacadeV13)

	reg("ApplicationOffers", 4, applicationoffers.NewOffersAPIV4)
	reg("ApplicationScaler", 1, applicationscaler.NewAPI)
	reg("Backups", 3, backups.NewFacadeV3)
	reg("Block", 2, block.NewAPI)
	reg("Bundle", 6, bundle.NewFacadeV6)
	reg("CharmDownloader", 1, charmdownloader.NewFacadeV1)
	reg("CharmRevisionUpdater", 2, charmrevisionupdater.NewCharmRevisionUpdaterAPI)
	reg("Charms", 4, charms.NewFacadeV4)
	reg("Cleaner", 2, cleaner.NewCleanerAPI)
	reg("Client", 4, client.NewFacade)
	reg("Cloud", 7, cloud.NewFacadeV7)

	// CAAS related facades.
	// Move these to the correct place above once the feature flag disappears.
	reg("CAASFirewaller", 1, caasfirewaller.NewStateFacadeLegacy)
	reg("CAASFirewallerEmbedded", 1, caasfirewaller.NewStateFacadeSidecar) // TODO(juju3): rename to CAASFirewallerSidecar
	reg("CAASOperator", 1, caasoperator.NewStateFacade)
	reg("CAASAdmission", 1, caasadmission.NewStateFacade)
	reg("CAASAgent", 2, caasagent.NewStateFacadeV2)
	reg("CAASModelOperator", 1, caasmodeloperator.NewAPIFromContext)
	reg("CAASOperatorProvisioner", 1, caasoperatorprovisioner.NewStateCAASOperatorProvisionerAPI)
	reg("CAASOperatorUpgrader", 1, caasoperatorupgrader.NewStateCAASOperatorUpgraderAPI)
	reg("CAASUnitProvisioner", 2, caasunitprovisioner.NewStateFacade)
	reg("CAASApplication", 1, caasapplication.NewStateFacade)
	reg("CAASApplicationProvisioner", 1, caasapplicationprovisioner.NewStateCAASApplicationProvisionerAPI)
	reg("CAASModelConfigManager", 1, caasmodelconfigmanager.NewFacade)

	reg("Controller", 11, controller.NewControllerAPIv11)
	reg("CrossModelRelations", 2, crossmodelrelations.NewStateCrossModelRelationsAPI)
	reg("CrossController", 1, crosscontroller.NewStateCrossControllerAPI)
	reg("CredentialManager", 1, credentialmanager.NewCredentialManagerAPI)
	reg("CredentialValidator", 2, credentialvalidator.NewCredentialValidatorAPI)
	reg("ExternalControllerUpdater", 1, externalcontrollerupdater.NewStateAPI)

	reg("Deployer", 1, deployer.NewDeployerAPI)
	reg("DiskManager", 2, diskmanager.NewDiskManagerAPI)
	reg("FanConfigurer", 1, fanconfigurer.NewFanConfigurerAPI)
	reg("Firewaller", 7, firewaller.NewFirewallerAPIV7)
	reg("FirewallRules", 1, firewallrules.NewFacade)
	reg("HighAvailability", 2, highavailability.NewHighAvailabilityAPI)
	reg("HostKeyReporter", 1, hostkeyreporter.NewFacade)
	reg("ImageManager", 2, imagemanager.NewImageManagerAPI)
	reg("ImageMetadata", 3, imagemetadata.NewAPI)

	reg("ImageMetadataManager", 1, imagemetadatamanager.NewAPI)

	reg("InstanceMutater", 2, instancemutater.NewFacadeV2)

	reg("InstancePoller", 4, instancepoller.NewFacade)
	reg("KeyManager", 1, keymanager.NewKeyManagerAPI)
	reg("KeyUpdater", 1, keyupdater.NewKeyUpdaterAPI)

	reg("LeadershipService", 2, leadership.NewLeadershipServiceFacade)

	reg("LifeFlag", 1, lifeflag.NewExternalFacade)
	reg("Logger", 1, loggerapi.NewLoggerAPI)
	reg("LogForwarding", 1, logfwd.NewFacade)
	reg("MachineActions", 1, machineactions.NewExternalFacade)

	reg("MachineManager", 6, machinemanager.NewFacadeV6)

	reg("MachineUndertaker", 1, machineundertaker.NewFacade)
	reg("Machiner", 5, machine.NewMachinerAPI)

	reg("MeterStatus", 2, meterstatus.NewMeterStatusFacade)
	reg("MetricsAdder", 2, metricsadder.NewMetricsAdderAPI)
	reg("MetricsDebug", 2, metricsdebug.NewMetricsDebugAPI)
	reg("MetricsManager", 1, metricsmanager.NewFacade)

	reg("MigrationFlag", 1, migrationflag.NewFacade)
	reg("MigrationMaster", 1, migrationmaster.NewMigrationMasterFacadeV1)
	reg("MigrationMaster", 2, migrationmaster.NewMigrationMasterFacadeV2)
	reg("MigrationMaster", 3, migrationmaster.NewMigrationMasterFacade) // Adds MinionReportTimeout.
	reg("MigrationMinion", 1, migrationminion.NewFacade)
	reg("MigrationTarget", 1, migrationtarget.NewFacade)

	reg("ModelConfig", 2, modelconfig.NewFacadeV2)
	reg("ModelGeneration", 4, modelgeneration.NewModelGenerationFacadeV4)
	reg("ModelManager", 9, modelmanager.NewFacadeV9)
	reg("ModelUpgrader", 1, modelupgrader.NewStateFacade)

	reg("Payloads", 1, payloads.NewFacade)
	regHookContext(
		"PayloadsHookContext", 1,
		payloadshookcontext.NewHookContextFacade,
		reflect.TypeOf(&payloadshookcontext.UnitFacade{}),
	)

	reg("Pinger", 1, NewPinger)
	reg("Provisioner", 11, provisioner.NewProvisionerAPIV11)

	reg("ProxyUpdater", 2, proxyupdater.NewFacadeV2)

	reg("RaftLease", 1, raftlease.NewFacadeV1)
	reg("RaftLease", 2, raftlease.NewFacadeV2)

	reg("Reboot", 2, reboot.NewRebootAPI)
	reg("RemoteRelations", 2, remoterelations.NewAPI)

	reg("Resources", 2, resources.NewFacadeV2)
	reg("ResourcesHookContext", 1, resourceshookcontext.NewStateFacade)

	reg("Resumer", 2, resumer.NewResumerAPI)
	reg("RetryStrategy", 1, retrystrategy.NewRetryStrategyAPI)
	reg("Singular", 2, singular.NewExternalFacade)
	reg("Secrets", 1, secrets.NewSecretsAPI)
	reg("SecretsManager", 1, secretsmanager.NewSecretManagerAPI)

	reg("SSHClient", 2, sshclient.NewFacade)

	reg("Spaces", 6, spaces.NewAPI)

	reg("StatusHistory", 2, statushistory.NewAPI)

	reg("Storage", 6, storage.NewStorageAPI) // modify Remove to support force and maxWait; add DetachStorage to support force and maxWait.

	reg("StorageProvisioner", 4, storageprovisioner.NewFacadeV4)
	reg("Subnets", 4, subnets.NewAPI) // Adds SubnetsByCIDR; removes AllSpaces.
	reg("Undertaker", 1, undertaker.NewUndertakerAPI)
	reg("UnitAssigner", 1, unitassigner.New)

	reg("Uniter", 18, uniter.NewUniterAPI)

	reg("Upgrader", 1, upgrader.NewUpgraderFacade)

	reg("UpgradeSeries", 3, upgradeseries.NewAPI)

	reg("UpgradeSteps", 2, upgradesteps.NewFacadeV2)
	reg("UserManager", 2, usermanager.NewUserManagerAPI)

	regRaw("AllWatcher", 1, NewAllWatcherV1, reflect.TypeOf((*SrvAllWatcherV1)(nil)))
	regRaw("AllWatcher", 2, NewAllWatcher, reflect.TypeOf((*SrvAllWatcher)(nil)))
	// Note: AllModelWatcher uses the same infrastructure as AllWatcher
	// but they are get under separate names as it possible the may
	// diverge in the future (especially in terms of authorisation
	// checks).
	regRaw("AllModelWatcher", 2, NewAllWatcherV1, reflect.TypeOf((*SrvAllWatcherV1)(nil)))
	regRaw("AllModelWatcher", 3, NewAllWatcher, reflect.TypeOf((*SrvAllWatcher)(nil)))
	regRaw("NotifyWatcher", 1, newNotifyWatcher, reflect.TypeOf((*srvNotifyWatcher)(nil)))
	regRaw("StringsWatcher", 1, newStringsWatcher, reflect.TypeOf((*srvStringsWatcher)(nil)))
	regRaw("OfferStatusWatcher", 1, newOfferStatusWatcher, reflect.TypeOf((*srvOfferStatusWatcher)(nil)))
	regRaw("RelationStatusWatcher", 1, newRelationStatusWatcher, reflect.TypeOf((*srvRelationStatusWatcher)(nil)))
	regRaw("RelationUnitsWatcher", 1, newRelationUnitsWatcher, reflect.TypeOf((*srvRelationUnitsWatcher)(nil)))
	regRaw("RemoteRelationWatcher", 1, newRemoteRelationWatcher, reflect.TypeOf((*srvRemoteRelationWatcher)(nil)))
	regRaw("VolumeAttachmentsWatcher", 2, newVolumeAttachmentsWatcher, reflect.TypeOf((*srvMachineStorageIdsWatcher)(nil)))
	regRaw("VolumeAttachmentPlansWatcher", 1, newVolumeAttachmentPlansWatcher, reflect.TypeOf((*srvMachineStorageIdsWatcher)(nil)))
	regRaw("FilesystemAttachmentsWatcher", 2, newFilesystemAttachmentsWatcher, reflect.TypeOf((*srvMachineStorageIdsWatcher)(nil)))
	regRaw("EntityWatcher", 2, newEntitiesWatcher, reflect.TypeOf((*srvEntitiesWatcher)(nil)))
	regRaw("MigrationStatusWatcher", 1, newMigrationStatusWatcher, reflect.TypeOf((*srvMigrationStatusWatcher)(nil)))
	regRaw("ModelSummaryWatcher", 1, newModelSummaryWatcher, reflect.TypeOf((*SrvModelSummaryWatcher)(nil)))
	regRaw("SecretsRotationWatcher", 1, newSecretsRotationWatcher, reflect.TypeOf((*srvSecretRotationWatcher)(nil)))

	return registry
}

// adminAPIFactories holds methods used to create
// admin APIs with specific versions.
var adminAPIFactories = map[int]adminAPIFactory{
	3: newAdminAPIV3,
}

// AdminFacadeDetails returns information on the Admin facade provided
// at login time. The Facade field of the returned slice elements will
// be nil.
func AdminFacadeDetails() []facade.Details {
	var fs []facade.Details
	for v, f := range adminAPIFactories {
		api := f(nil, nil, nil)
		t := reflect.TypeOf(api)
		fs = append(fs, facade.Details{
			Name:    "Admin",
			Version: v,
			Type:    t,
		})
	}
	return fs
}

type hookContextFacadeFn func(*state.State, *state.Unit) (interface{}, error)

// regHookContextFacade registers facades for use within a hook
// context. This function handles the translation from a
// hook-context-facade to a standard facade so the caller's factory
// method can elide unnecessary arguments. This function also handles
// any necessary authorization for the client.
//
// XXX(fwereade): this is fundamentally broken, because it (1)
// arbitrarily creates a new facade for a tiny fragment of a specific
// client worker's reponsibilities and (2) actively conceals necessary
// auth information from the facade. Don't call it; actively work to
// delete code that uses it, and rewrite it properly.
func regHookContextFacade(
	reg *facade.Registry,
	name string,
	version int,
	newHookContextFacade hookContextFacadeFn,
	facadeType reflect.Type,
) error {
	newFacade := func(context facade.Context) (facade.Facade, error) {
		authorizer := context.Auth()
		st := context.State()

		if !authorizer.AuthUnitAgent() {
			return nil, apiservererrors.ErrPerm
		}
		// Verify that the unit's ID matches a unit that we know about.
		tag := authorizer.GetAuthTag()
		if _, ok := tag.(names.UnitTag); !ok {
			return nil, errors.Errorf("expected names.UnitTag, got %T", tag)
		}
		unit, err := st.Unit(tag.Id())
		if err != nil {
			return nil, errors.Trace(err)
		}
		return newHookContextFacade(st, unit)
	}
	err := reg.Register(name, version, newFacade, facadeType)
	return errors.Trace(err)
}
