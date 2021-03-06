// Copyright 2016 Canonical Ltd.  All rights reserved.

// Package updateallocation defines the command used to update allocations.
package updateallocation

import (
	"fmt"
	"strconv"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/juju/cmd/modelcmd"
	"github.com/juju/juju/environs/configstore"
	"gopkg.in/macaroon-bakery.v1/httpbakery"
	"launchpad.net/gnuflag"

	api "github.com/juju/romulus/api/budget"
	rcmd "github.com/juju/romulus/cmd"
)

type updateAllocationCommand struct {
	modelcmd.ModelCommandBase
	rcmd.HttpCommand
	Name  string
	Value string
}

// NewUpdateAllocationCommand returns a new updateAllocationCommand.
func NewUpdateAllocationCommand() cmd.Command {
	return modelcmd.Wrap(&updateAllocationCommand{})
}

var newAPIClient = func(c *httpbakery.Client) (apiClient, error) {
	client := api.NewClient(c)
	return client, nil
}

type apiClient interface {
	UpdateAllocation(string, string, string) (string, error)
}

const doc = `
Updates an existing allocation on a service.

Example:
 juju update-allocation wordpress 10
     Sets the allocation for the wordpress service to 10.
`

// Info implements cmd.Command.Info.
func (c *updateAllocationCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "update-allocation",
		Purpose: "update an allocation",
		Doc:     doc,
	}
}

// SetFlags implements cmd.Command.
func (c *updateAllocationCommand) SetFlags(f *gnuflag.FlagSet) {
	c.ModelCommandBase.SetFlags(f)
}

// AllowInterspersed implements cmd.Command.
func (c *updateAllocationCommand) AllowInterspersedFlags() bool { return true }

// IsSuperCommand implements cmd.Command.
// Defined here because of ambiguity between HttpCommand and ModelCommandBase.
func (c *updateAllocationCommand) IsSuperCommand() bool { return false }

// Init implements cmd.Command.Init.
func (c *updateAllocationCommand) Init(args []string) error {
	if len(args) < 2 {
		return errors.New("service and value required")
	}
	c.Name, c.Value = args[0], args[1]
	if _, err := strconv.ParseInt(c.Value, 10, 32); err != nil {
		return errors.New("value needs to be a whole number")
	}
	return cmd.CheckEmpty(args[2:])
}

func (c *updateAllocationCommand) modelUUID() (string, error) {
	store, err := configstore.Default()
	if err != nil {
		return "", errors.Trace(err)
	}
	modelInfo, err := store.ReadInfo(c.ModelName())
	if err != nil {
		return "", errors.Trace(err)
	}
	return modelInfo.APIEndpoint().ModelUUID, nil
}

// Run implements cmd.Command.Run and contains most of the setbudget logic.
func (c *updateAllocationCommand) Run(ctx *cmd.Context) error {
	defer c.Close()
	modelUUID, err := c.modelUUID()
	if err != nil {
		return errors.Annotate(err, "failed to get model uuid")
	}
	client, err := c.NewClient()
	if err != nil {
		return errors.Annotate(err, "failed to create an http client")
	}
	api, err := newAPIClient(client)
	if err != nil {
		return errors.Annotate(err, "failed to create an api client")
	}
	resp, err := api.UpdateAllocation(modelUUID, c.Name, c.Value)
	if err != nil {
		return errors.Annotate(err, "failed to update the allocation")
	}
	fmt.Fprintf(ctx.Stdout, resp)
	return nil
}
