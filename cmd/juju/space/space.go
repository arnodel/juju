// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package space

import (
	"io"
	"net"

	"github.com/juju/cmd"
	"github.com/juju/collections/set"
	"github.com/juju/errors"
	"github.com/juju/loggo"
	"gopkg.in/juju/names.v3"

	"github.com/juju/juju/api"
	"github.com/juju/juju/api/spaces"
	"github.com/juju/juju/api/subnets"
	"github.com/juju/juju/apiserver/params"
	"github.com/juju/juju/cmd/modelcmd"
	"github.com/juju/juju/core/network"
)

// SpaceAPI defines the necessary API methods needed by the space
// subcommands.
type SpaceAPI interface {
	// ListSpaces returns all Juju network spaces and their subnets.
	ListSpaces() ([]params.Space, error)

	// AddSpace adds a new Juju network space, associating the
	// specified subnets with it (optional; can be empty), setting the
	// space and subnets access to public or private.
	AddSpace(name string, subnetIds []string, public bool) error

	// TODO(dimitern): All of the following api methods should take
	// names.SpaceTag instead of name, the only exceptions are
	// AddSpace, and RenameSpace as the named space doesn't exist
	// yet.

	// RemoveSpace removes an existing Juju network space, transferring
	// any associated subnets to the default space.
	RemoveSpace(name string, force bool, dryRun bool) (network.RemoveSpace, error)

	// RenameSpace changes the name of the space.
	RenameSpace(name, newName string) error

	// ReloadSpaces fetches spaces and subnets from substrate
	ReloadSpaces() error

	// ShowSpace fetches space information.
	ShowSpace(name string) (network.ShowSpace, error)

	// MoveSubnets ensures that the input subnets are in the input space.
	MoveSubnets(names.SpaceTag, []names.SubnetTag, bool) (params.MoveSubnetsResult, error)
}

// SubnetAPI defines the necessary API methods needed by the subnet subcommands.
type SubnetAPI interface {

	// SubnetsByCIDR returns the collection of subnets matching each CIDR in the input.
	SubnetsByCIDR([]string) ([]params.SubnetsResult, error)
}

// API defines the contract for requesting the API facades.
type API interface {
	io.Closer

	SpaceAPI
	SubnetAPI
}

var logger = loggo.GetLogger("juju.cmd.juju.space")

// SpaceCommandBase is the base type embedded into all space subcommands.
type SpaceCommandBase struct {
	modelcmd.ModelCommandBase
	modelcmd.IAASOnlyCommand
	api API
}

// ParseNameAndCIDRs verifies the input args and returns any errors,
// like missing/invalid name or CIDRs (validated when given, but it's
// an error for CIDRs to be empty if cidrsOptional is false).
func ParseNameAndCIDRs(args []string, cidrsOptional bool) (
	name string, CIDRs set.Strings, err error,
) {
	defer errors.DeferredAnnotatef(&err, "invalid arguments specified")

	if len(args) == 0 {
		return "", nil, errors.New("space name is required")
	}
	name, err = CheckName(args[0])
	if err != nil {
		return name, nil, errors.Trace(err)
	}

	CIDRs, err = CheckCIDRs(args[1:], cidrsOptional)
	return name, CIDRs, errors.Trace(err)
}

// CheckName checks whether name is a valid space name.
func CheckName(name string) (string, error) {
	// Validate given name.
	if !names.IsValidSpace(name) {
		return "", errors.Errorf("%q is not a valid space name", name)
	}
	return name, nil
}

// CheckCIDRs parses the list of strings as CIDRs, checking for
// correct formatting, no duplication and no overlaps. Returns error
// if no CIDRs are provided, unless cidrsOptional is true.
func CheckCIDRs(args []string, cidrsOptional bool) (set.Strings, error) {
	// Validate any given CIDRs.
	CIDRs := set.NewStrings()
	for _, arg := range args {
		_, ipNet, err := net.ParseCIDR(arg)
		if err != nil {
			logger.Debugf("cannot parse %q: %v", arg, err)
			return CIDRs, errors.Errorf("%q is not a valid CIDR", arg)
		}
		cidr := ipNet.String()
		if CIDRs.Contains(cidr) {
			if cidr == arg {
				return CIDRs, errors.Errorf("duplicate subnet %q specified", cidr)
			}
			return CIDRs, errors.Errorf("subnet %q overlaps with %q", arg, cidr)
		}
		CIDRs.Add(cidr)
	}

	if CIDRs.IsEmpty() && !cidrsOptional {
		return CIDRs, errors.New("CIDRs required but not provided")
	}

	return CIDRs, nil
}

// APIShim forwards SpaceAPI methods to the real API facade for
// implemented methods only.
type APIShim struct {
	SpaceAPI

	apiState  api.Connection
	spaceAPI  *spaces.API
	subnetAPI *subnets.API
}

func (m *APIShim) Close() error {
	return m.apiState.Close()
}

// AddSpace adds a new Juju network space, associating the
// specified subnets with it (optional; can be empty), setting the
// space and subnets access to public or private.
func (m *APIShim) AddSpace(name string, subnetIds []string, public bool) error {
	return m.spaceAPI.CreateSpace(name, subnetIds, public)
}

// ListSpaces returns all Juju network spaces and their subnets.
func (m *APIShim) ListSpaces() ([]params.Space, error) {
	return m.spaceAPI.ListSpaces()
}

// ReloadSpaces fetches spaces and subnets from substrate
func (m *APIShim) ReloadSpaces() error {
	return m.spaceAPI.ReloadSpaces()
}

// RemoveSpace removes an existing Juju network space, transferring
// any associated subnets to the default space.
func (m *APIShim) RemoveSpace(name string, force bool, dryRun bool) (network.RemoveSpace, error) {
	return m.spaceAPI.RemoveSpace(name, force, dryRun)
}

// RenameSpace changes the name of the space.
func (m *APIShim) RenameSpace(oldName, newName string) error {
	return m.spaceAPI.RenameSpace(oldName, newName)
}

// ShowSpace fetches space information.
func (m *APIShim) ShowSpace(name string) (network.ShowSpace, error) {
	return m.spaceAPI.ShowSpace(name)
}

// MoveSubnets ensures that the input subnets are in the input space.
func (m *APIShim) MoveSubnets(space names.SpaceTag, subnets []names.SubnetTag, force bool) (params.MoveSubnetsResult, error) {
	return m.spaceAPI.MoveSubnets(space, subnets, force)
}

// SubnetsByCIDR returns the collection of subnets matching each CIDR in the input.
func (m *APIShim) SubnetsByCIDR(cidrs []string) ([]params.SubnetsResult, error) {
	return m.subnetAPI.SubnetsByCIDR(cidrs)
}

// NewAPI returns a API for the root api endpoint that the
// environment command returns.
func (c *SpaceCommandBase) NewAPI() (API, error) {
	if c.api != nil {
		// Already added.
		return c.api, nil
	}
	root, err := c.NewAPIRoot()
	if err != nil {
		return nil, errors.Trace(err)
	}

	// This is tested with a feature test.
	shim := &APIShim{
		apiState:  root,
		spaceAPI:  spaces.NewAPI(root),
		subnetAPI: subnets.NewAPI(root),
	}
	return shim, nil
}

type RunOnAPI func(api API, ctx *cmd.Context) error

func (c *SpaceCommandBase) RunWithAPI(ctx *cmd.Context, toRun RunOnAPI) error {
	api, err := c.NewAPI()
	if err != nil {
		return errors.Annotate(err, "cannot connect to the API server")
	}
	defer api.Close()
	return toRun(api, ctx)
}

type RunOnSpaceAPI func(api SpaceAPI, ctx *cmd.Context) error

func (c *SpaceCommandBase) RunWithSpaceAPI(ctx *cmd.Context, toRun RunOnSpaceAPI) error {
	api, err := c.NewAPI()
	if err != nil {
		return errors.Annotate(err, "cannot connect to the API server")
	}
	defer api.Close()
	return toRun(api, ctx)
}
