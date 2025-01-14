// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package subnets

import (
	"github.com/juju/collections/set"
	"github.com/juju/errors"
	"github.com/juju/loggo"
	"github.com/juju/names/v4"

	"github.com/juju/juju/apiserver/common/networkingcommon"
	apiservererrors "github.com/juju/juju/apiserver/errors"
	"github.com/juju/juju/apiserver/facade"
	"github.com/juju/juju/apiserver/params"
	"github.com/juju/juju/core/network"
	"github.com/juju/juju/core/permission"
	"github.com/juju/juju/environs"
	"github.com/juju/juju/environs/context"
	"github.com/juju/juju/state"
)

var logger = loggo.GetLogger("juju.apiserver.subnets")

// Backing contains the state methods used in this package.
type Backing interface {
	environs.EnvironConfigGetter

	// AvailabilityZones returns all cached availability zones (i.e.
	// not from the provider, but in state).
	AvailabilityZones() (network.AvailabilityZones, error)

	// SetAvailabilityZones replaces the cached list of availability
	// zones with the given zones.
	SetAvailabilityZones(network.AvailabilityZones) error

	// AddSubnet creates a backing subnet for an existing subnet.
	AddSubnet(networkingcommon.BackingSubnetInfo) (networkingcommon.BackingSubnet, error)

	// AllSubnets returns all backing subnets.
	AllSubnets() ([]networkingcommon.BackingSubnet, error)

	// SubnetByCIDR returns a unique subnet based on the input CIDR.
	SubnetByCIDR(cidr string) (networkingcommon.BackingSubnet, error)

	// SubnetsByCIDR returns any subnets with the input CIDR.
	SubnetsByCIDR(cidr string) ([]networkingcommon.BackingSubnet, error)

	// AllSpaces returns all known Juju network spaces.
	AllSpaces() ([]networkingcommon.BackingSpace, error)

	// ModelTag returns the tag of the model this state is associated to.
	ModelTag() names.ModelTag
}

// API provides the subnets API facade for version 4.
type API struct {
	backing    Backing
	resources  facade.Resources
	authorizer facade.Authorizer
	context    context.ProviderCallContext
}

// NewAPI creates a new Subnets API server-side facade with a
// state.State backing.
func NewAPI(st *state.State, res facade.Resources, auth facade.Authorizer) (*API, error) {
	stateShim, err := NewStateShim(st)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return newAPIWithBacking(stateShim, context.CallContext(st), res, auth)
}

func (api *API) checkCanRead() error {
	canRead, err := api.authorizer.HasPermission(permission.ReadAccess, api.backing.ModelTag())
	if err != nil {
		return errors.Trace(err)
	}
	if !canRead {
		return apiservererrors.ServerError(apiservererrors.ErrPerm)
	}
	return nil
}

func (api *API) checkCanWrite() error {
	canWrite, err := api.authorizer.HasPermission(permission.WriteAccess, api.backing.ModelTag())
	if err != nil {
		return errors.Trace(err)
	}
	if !canWrite {
		return apiservererrors.ServerError(apiservererrors.ErrPerm)
	}
	return nil
}

// newAPIWithBacking creates a new server-side Subnets API facade with
// a common.NetworkBacking
func newAPIWithBacking(backing Backing, ctx context.ProviderCallContext, resources facade.Resources, authorizer facade.Authorizer) (*API, error) {
	// Only clients can access the Subnets facade.
	if !authorizer.AuthClient() {
		return nil, apiservererrors.ErrPerm
	}
	return &API{
		backing:    backing,
		resources:  resources,
		authorizer: authorizer,
		context:    ctx,
	}, nil
}

// AllZones returns all availability zones known to Juju. If a
// zone is unusable, unavailable, or deprecated the Available
// field will be false.
func (api *API) AllZones() (params.ZoneResults, error) {
	if err := api.checkCanRead(); err != nil {
		return params.ZoneResults{}, err
	}
	return allZones(api.context, api.backing)
}

// AddSubnets adds existing subnets to Juju.
func (api *API) AddSubnets(args params.AddSubnetsParams) (params.ErrorResults, error) {
	if err := api.checkCanWrite(); err != nil {
		return params.ErrorResults{}, err
	}

	results := params.ErrorResults{
		Results: make([]params.ErrorResult, len(args.Subnets)),
	}

	if len(args.Subnets) == 0 {
		return results, nil
	}

	cache := NewAddSubnetsCache(api.backing)
	for i, arg := range args.Subnets {
		err := addOneSubnet(api.context, api.backing, arg, cache)
		if err != nil {
			results.Results[i].Error = apiservererrors.ServerError(err)
		}
	}
	return results, nil
}

// ListSubnets returns the matching subnets after applying
// optional filters.
func (api *API) ListSubnets(args params.SubnetsFilters) (results params.ListSubnetsResults, err error) {
	if err := api.checkCanRead(); err != nil {
		return params.ListSubnetsResults{}, err
	}

	subs, err := api.backing.AllSubnets()
	if err != nil {
		return results, errors.Trace(err)
	}

	var spaceFilter string
	if args.SpaceTag != "" {
		tag, err := names.ParseSpaceTag(args.SpaceTag)
		if err != nil {
			return results, errors.Trace(err)
		}
		spaceFilter = tag.Id()
	}
	zoneFilter := args.Zone

	for _, subnet := range subs {
		if spaceFilter != "" && subnet.SpaceName() != spaceFilter {
			logger.Tracef(
				"filtering subnet %q from space %q not matching filter %q",
				subnet.CIDR(), subnet.SpaceName(), spaceFilter,
			)
			continue
		}
		zoneSet := set.NewStrings(subnet.AvailabilityZones()...)
		if zoneFilter != "" && !zoneSet.IsEmpty() && !zoneSet.Contains(zoneFilter) {
			logger.Tracef(
				"filtering subnet %q with zones %v not matching filter %q",
				subnet.CIDR(), subnet.AvailabilityZones(), zoneFilter,
			)
			continue
		}

		results.Results = append(results.Results, networkingcommon.BackingSubnetToParamsSubnet(subnet))
	}
	return results, nil
}

// SubnetsByCIDR returns the collection of subnets matching each CIDR in the input.
func (api *API) SubnetsByCIDR(arg params.CIDRParams) (params.SubnetsResults, error) {
	result := params.SubnetsResults{}

	if err := api.checkCanRead(); err != nil {
		return result, err
	}

	results := make([]params.SubnetsResult, len(arg.CIDRS))
	for i, cidr := range arg.CIDRS {
		if !network.IsValidCIDR(cidr) {
			results[i].Error = apiservererrors.ServerError(errors.NotValidf("CIDR %q", cidr))
			continue
		}

		subnets, err := api.backing.SubnetsByCIDR(cidr)
		if err != nil {
			results[i].Error = apiservererrors.ServerError(err)
			continue
		}

		subnetResults := make([]params.SubnetV2, len(subnets))
		for j, subnet := range subnets {
			subnetResults[j] = networkingcommon.BackingSubnetToParamsSubnetV2(subnet)
		}
		results[i].Subnets = subnetResults
	}
	result.Results = results
	return result, nil
}
