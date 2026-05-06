package services

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/Azure/terraform-provider-azapi/internal/clients"
	"github.com/Azure/terraform-provider-azapi/internal/docstrings"
	"github.com/Azure/terraform-provider-azapi/internal/locks"
	"github.com/Azure/terraform-provider-azapi/internal/services/common"
	"github.com/Azure/terraform-provider-azapi/internal/services/myvalidator"
	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	tffwdocs "github.com/magodo/terraform-plugin-framework-docs"
)

// AzapiResourceActionModel defines the configuration for the action equivalent of the stateful `azapi_resource_action` resource.
type AzapiResourceActionModel struct {
	Type            types.String  `tfsdk:"type"`
	ResourceId      types.String  `tfsdk:"resource_id"`
	Action          types.String  `tfsdk:"action"`
	Method          types.String  `tfsdk:"method"`
	Body            types.Dynamic `tfsdk:"body"`
	Locks           types.List    `tfsdk:"locks"`
	Headers         types.Map     `tfsdk:"headers"`
	QueryParameters types.Map     `tfsdk:"query_parameters"`
}

// AzapiResourceAction implements performing an action without persisting state (stateless invocation).
type AzapiResourceAction struct {
	ProviderData *clients.Client
}

var _ action.Action = &AzapiResourceAction{}
var _ action.ActionWithConfigure = &AzapiResourceAction{}
var _ tffwdocs.ActionWithRenderOption = &AzapiResourceAction{}

func (a *AzapiResourceAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_resource_action", req.ProviderTypeName)
}

func (a *AzapiResourceAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if v, ok := req.ProviderData.(*clients.Client); ok {
		a.ProviderData = v
	}
}

func (a *AzapiResourceAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = actionschema.Schema{
		Description: "Perform an action on an existing Azure resource (stateless).",
		Attributes: map[string]actionschema.Attribute{
			"type": actionschema.StringAttribute{
				Required:            true,
				MarkdownDescription: docstrings.Type(),
				Validators: []validator.String{
					myvalidator.StringIsResourceType(),
				},
			},

			"resource_id": actionschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of an existing Azure resource.",
				Validators: []validator.String{
					myvalidator.StringIsResourceID(),
				},
			},

			"action": actionschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: docstrings.ResourceAction(),
			},

			"method": actionschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "HTTP method for the action. Defaults to POST.",
				Validators: []validator.String{
					stringvalidator.OneOf("POST", "PATCH", "PUT", "DELETE", "GET", "HEAD"),
				},
			},

			"body": actionschema.DynamicAttribute{
				Optional:            true,
				MarkdownDescription: docstrings.Body(),
				Validators: []validator.Dynamic{
					myvalidator.DynamicIsNotStringValidator(),
				},
			},

			"locks": actionschema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: docstrings.Locks(),
				Validators: []validator.List{
					listvalidator.ValueStringsAre(myvalidator.StringIsNotEmpty()),
				},
			},

			"headers": actionschema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Headers to include in the request.",
			},

			"query_parameters": actionschema.MapAttribute{
				Optional:            true,
				ElementType:         types.ListType{ElemType: types.StringType},
				MarkdownDescription: "Query parameters to include in the request.",
			},
		},
	}
}

func (a *AzapiResourceAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config AzapiResourceActionModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	method := config.Method.ValueString()
	if method == "" {
		method = "POST"
	}

	id, err := parse.ResourceIDWithResourceType(config.ResourceId.ValueString(), config.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid configuration", err.Error())
		return
	}
	ctx = tflog.SetField(ctx, "resource_id", id.ID())

	var requestBody interface{}
	if err := unmarshalBody(config.Body, &requestBody); err != nil {
		resp.Diagnostics.AddError("Invalid body", fmt.Sprintf("The argument \"body\" is invalid: %s", err.Error()))
		return
	}

	lockIds := common.AsStringList(config.Locks)
	slices.Sort(lockIds)
	for _, lockId := range lockIds {
		locks.ByID(lockId)
		defer locks.UnlockByID(lockId)
	}

	requestOptions := clients.RequestOptions{
		Headers:         common.AsMapOfString(config.Headers),
		QueryParameters: clients.NewQueryParameters(common.AsMapOfLists(config.QueryParameters)),
	}

	// Default timeout similar to resource create
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	if resp.SendProgress != nil {
		resp.SendProgress(action.InvokeProgressEvent{Message: "Invoking Azure resource action"})
	}

	_, err = a.ProviderData.ResourceClient.Action(ctx, id.AzureResourceId, config.Action.ValueString(), id.ApiVersion, method, requestBody, requestOptions)
	if err != nil {
		resp.Diagnostics.AddError("Failed to perform action", fmt.Sprintf("performing action %s of %q: %+v", config.Action.ValueString(), id, err))
		return
	}

	if resp.SendProgress != nil {
		resp.SendProgress(action.InvokeProgressEvent{Message: "Action completed"})
	}
}

func (a *AzapiResourceAction) RenderOption() tffwdocs.ActionRenderOption {
	return tffwdocs.ActionRenderOption{
		Examples: []tffwdocs.Example{
			{
				HCL: `
terraform {
  required_providers {
    azapi = {
      source = "Azure/azapi"
    }
  }
}

provider "azapi" {}

variable "admin_password" {
  type      = string
  sensitive = true
}

resource "azapi_resource" "resource_group" {
  type     = "Microsoft.Resources/resourceGroups@2020-06-01"
  name     = "example-resources"
  location = "West Europe"
}

resource "azapi_resource" "virtual_network" {
  type      = "Microsoft.Network/virtualNetworks@2023-11-01"
  name      = "example-network"
  parent_id = azapi_resource.resource_group.id
  location  = azapi_resource.resource_group.location
  body = {
    properties = {
      addressSpace = {
        addressPrefixes = ["10.0.0.0/16"]
      }
    }
  }
}

resource "azapi_resource" "subnet" {
  type      = "Microsoft.Network/virtualNetworks/subnets@2023-11-01"
  name      = "internal"
  parent_id = azapi_resource.virtual_network.id
  body = {
    properties = {
      addressPrefix = "10.0.2.0/24"
    }
  }
}

resource "azapi_resource" "network_interface" {
  type      = "Microsoft.Network/networkInterfaces@2023-11-01"
  name      = "example-nic"
  parent_id = azapi_resource.resource_group.id
  location  = azapi_resource.resource_group.location
  body = {
    properties = {
      ipConfigurations = [
        {
          name = "internal"
          properties = {
            subnet = {
              id = azapi_resource.subnet.id
            }
            privateIPAllocationMethod = "Dynamic"
          }
        }
      ]
    }
  }
}

resource "azapi_resource" "windows_virtual_machine" {
  type      = "Microsoft.Compute/virtualMachines@2023-03-01"
  name      = "example-vm"
  parent_id = azapi_resource.resource_group.id
  location  = azapi_resource.resource_group.location
  body = {
    properties = {
      hardwareProfile = {
        vmSize = "Standard_F2"
      }
      osProfile = {
        computerName  = "examplevm"
        adminUsername = "adminuser"
        adminPassword = var.admin_password
      }
      networkProfile = {
        networkInterfaces = [
          {
            id = azapi_resource.network_interface.id
          }
        ]
      }
      storageProfile = {
        imageReference = {
          publisher = "MicrosoftWindowsServer"
          offer     = "WindowsServer"
          sku       = "2016-Datacenter"
          version   = "latest"
        }
        osDisk = {
          caching      = "ReadWrite"
          createOption = "FromImage"
          managedDisk = {
            storageAccountType = "Standard_LRS"
          }
        }
      }
    }
  }
  tags = {
    environment = "production"
  }

  lifecycle {
    action_trigger {
      events  = [before_update]
      actions = [action.azapi_resource_action.power_off]
    }

    action_trigger {
      events  = [after_update]
      actions = [action.azapi_resource_action.power_on]
    }
  }
}

action "azapi_resource_action" "power_off" {
  config {
    type        = "Microsoft.Compute/virtualMachines@2023-03-01"
    resource_id = azapi_resource.windows_virtual_machine.id
    action      = "powerOff"
  }
}

action "azapi_resource_action" "power_on" {
  config {
    type        = "Microsoft.Compute/virtualMachines@2023-03-01"
    resource_id = azapi_resource.windows_virtual_machine.id
    action      = "start"
  }
}`,
			},
		},
	}
}
