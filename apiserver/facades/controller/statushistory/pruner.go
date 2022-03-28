// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package statushistory

import (
	"github.com/juju/juju/apiserver/common"
	apiservererrors "github.com/juju/juju/apiserver/errors"
	"github.com/juju/juju/apiserver/facade"
	"github.com/juju/juju/rpc/params"
	"github.com/juju/juju/state"
)

// API is the concrete implementation of the Pruner endpoint.
type API struct {
	*common.ModelWatcher
	st         *state.State
	authorizer facade.Authorizer
}

// NewAPI returns an API Instance.
func NewAPI(ctx facade.Context) (*API, error) {
	st := ctx.State()
	m, err := st.Model()
	if err != nil {
		return nil, err
	}

	auth := ctx.Auth()
	return &API{
		ModelWatcher: common.NewModelWatcher(m, ctx.Resources(), auth),
		st:           st,
		authorizer:   auth,
	}, nil
}

// Prune endpoint removes status history entries until
// only the ones newer than now - p.MaxHistoryTime remain and
// the history is smaller than p.MaxHistoryMB.
func (api *API) Prune(p params.StatusHistoryPruneArgs) error {
	if !api.authorizer.AuthController() {
		return apiservererrors.ErrPerm
	}
	return state.PruneStatusHistory(api.st, p.MaxHistoryTime, p.MaxHistoryMB)
}
