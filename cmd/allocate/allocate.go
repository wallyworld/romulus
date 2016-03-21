// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package allocate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/juju/cmd/modelcmd"
	"gopkg.in/macaroon-bakery.v1/httpbakery"

	api "github.com/juju/romulus/api/budget"
)

var budgetWithLimitRe = regexp.MustCompile(`^[a-zA-Z0-9\-]+:[1-9][0-9]*$`)

type allocateCommand struct {
	modelcmd.ModelCommandBase
	api      apiClient
	Budget   string
	Model    string
	Services []string
	Limit    string
}

// NewAllocateCommand returns a new allocateCommand
func NewAllocateCommand() modelcmd.ModelCommand {
	return &allocateCommand{}
}

const doc = `
Allocate budget for the specified services, replacing any prior allocations
made for the specified services.

Usage:

 juju allocate <budget>:<value> <service> [<service2> ...]

Example:

 juju allocate somebudget:42 db
     Assigns service "db" to an allocation on budget "somebudget" with the limit "42".
`

// Info implements cmd.Command.Info.
func (c *allocateCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "allocate",
		Purpose: "allocate budget to services",
		Doc:     doc,
	}
}

// Init implements cmd.Command.Init.
func (c *allocateCommand) Init(args []string) error {
	if len(args) < 2 {
		return errors.New("budget and service name required")
	}
	budgetWithLimit := args[0]
	var err error
	c.Budget, c.Limit, err = parseBudgetWithLimit(budgetWithLimit)
	if err != nil {
		return err
	}
	c.Model, err = c.modelUUID()
	if err != nil {
		return err
	}

	c.Services = args[1:]
	return nil
}

// Run implements cmd.Command.Run and has most of the logic for the run command.
func (c *allocateCommand) Run(ctx *cmd.Context) error {
	client, err := c.BakeryClient()
	if err != nil {
		return errors.Annotate(err, "failed to create an http client")
	}
	api, err := c.newAPIClient(client)
	if err != nil {
		return errors.Annotate(err, "failed to create an api client")
	}
	resp, err := api.CreateAllocation(c.Budget, c.Limit, c.Model, c.Services)
	if err != nil {
		return errors.Annotate(err, "failed to create allocation")
	}
	fmt.Fprintf(ctx.Stdout, resp)
	return nil
}

func (c *allocateCommand) modelUUID() (string, error) {
	model, err := c.ClientStore().ModelByName(c.ControllerName(), c.AccountName(), c.ModelName())
	if err != nil {
		return "", errors.Trace(err)
	}
	return model.ModelUUID, nil
}

func parseBudgetWithLimit(bl string) (string, string, error) {
	if !budgetWithLimitRe.MatchString(bl) {
		return "", "", errors.New("invalid budget specification, expecting <budget>:<limit>")
	}
	parts := strings.Split(bl, ":")
	return parts[0], parts[1], nil
}

func (c *allocateCommand) newAPIClient(bakery *httpbakery.Client) (apiClient, error) {
	if c.api != nil {
		return c.api, nil
	}
	c.api = api.NewClient(bakery)
	return c.api, nil
}

type apiClient interface {
	CreateAllocation(string, string, string, []string) (string, error)
}
