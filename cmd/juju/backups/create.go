// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package backups

import (
	"fmt"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"launchpad.net/gnuflag"

	"github.com/juju/juju/apiserver/params"
)

const createDoc = `
"create" requests that juju create a backup of its state and print the
backup's unique ID.  You may provide a note to associate with the backup.

The backup archive and associated metadata are stored in juju and
will be lost when the environment is destroyed.
`

var sendCreateRequest = func(cmd *BackupsCreateCommand) (*params.BackupsMetadataResult, error) {
	client, err := cmd.NewAPIClient()
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer client.Close()

	return client.Create(cmd.Notes)
}

// BackupsCreateCommand is the sub-command for creating a new backup.
type BackupsCreateCommand struct {
	BackupsCommandBase
	// Quiet indicates that the full metadata should not be dumped.
	Quiet bool
	// Notes is the custom message to associated with the new backup.
	Notes string
}

// Info implements Command.Info.
func (c *BackupsCreateCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "create",
		Args:    "[<notes>]",
		Purpose: "create a backup",
		Doc:     createDoc,
	}
}

// SetFlags implements Command.SetFlags.
func (c *BackupsCreateCommand) SetFlags(f *gnuflag.FlagSet) {
	f.BoolVar(&c.Quiet, "quiet", false, "do not print the metadata")
}

// Init implements Command.Init.
func (c *BackupsCreateCommand) Init(args []string) error {
	notes, err := cmd.ZeroOrOneArgs(args)
	if err != nil {
		return err
	}
	c.Notes = notes
	return nil
}

// Run implements Command.Run.
func (c *BackupsCreateCommand) Run(ctx *cmd.Context) error {
	result, err := sendCreateRequest(c)
	if err != nil {
		return errors.Trace(err)
	}

	if !c.Quiet {
		c.dumpMetadata(ctx, result)
	}

	fmt.Fprintln(ctx.Stdout, result.ID)
	return nil
}
