---
subcategory: "Microsoft.StandbyPool - Standby Pools"
page_title: "standbyvirtualmachinepools"
description: |-
  Manages a Microsoft Standby pools for Virtual Machine Scale Sets.
---

# Microsoft.StandbyPool/standbyvirtualmachinepools - Microsoft Standby pools for Virtual Machine Scale Sets

This article demonstrates how to use `azapi` provider to manage the Microsoft Standby pools for Virtual Machine Scale Sets resource in Azure.



## Example Usage

### default

```hcl
terraform {
  required_providers {
    azapi = {
      source = "Azure/azapi"
    }
  }
}

provider "azapi" {
  skip_provider_registration = false
}

variable "resource_name" {
  type    = string
  default = "acctest0001"
}

variable "location" {
  type    = string
  default = "eastus"
}


resource "azapi_resource" "resourceGroup" {
  type     = "Microsoft.Resources/resourceGroups@2020-06-01"
  name     = "${var.resource_name}-rg"
  location = var.location
}

resource "azapi_resource" "virtualNetwork" {
  type      = "Microsoft.Network/virtualNetworks@2022-07-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = "${var.resource_name}-vnet"
  location  = var.location
  body = {
    properties = {
      addressSpace = {
        addressPrefixes = [
          "10.0.0.0/16",
        ]
      }
      dhcpOptions = {
        dnsServers = [
        ]
      }
      subnets = [
      ]
    }
  }
  schema_validation_enabled = false
  response_export_values    = ["*"]
  lifecycle {
    ignore_changes = [body.properties.subnets]
  }
}

resource "azapi_resource" "subnet" {
  type      = "Microsoft.Network/virtualNetworks/subnets@2022-07-01"
  parent_id = azapi_resource.virtualNetwork.id
  name      = "${var.resource_name}-subnet"
  body = {
    properties = {
      addressPrefix = "10.0.2.0/24"
      delegations = [
      ]
      privateEndpointNetworkPolicies    = "Enabled"
      privateLinkServiceNetworkPolicies = "Enabled"
      serviceEndpointPolicies = [
      ]
      serviceEndpoints = [
      ]
    }
  }
  schema_validation_enabled = false
  response_export_values    = ["*"]
}

resource "azapi_resource" "virtualMachineScaleSet" {
  type      = "Microsoft.Compute/virtualMachineScaleSets@2023-03-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = "${var.resource_name}-vmss"
  location  = var.location
  body = {
    properties = {
      additionalCapabilities = {
      }
      orchestrationMode        = "Flexible"
      platformFaultDomainCount = 1
      singlePlacementGroup     = false
      upgradePolicy = {
        mode = "Manual"
      }
      virtualMachineProfile = {
        diagnosticsProfile = {
          bootDiagnostics = {
            enabled    = false
            storageUri = ""
          }
        }
        extensionProfile = {
          extensionsTimeBudget = "PT1H30M"
        }
        networkProfile = {
          networkApiVersion = "2022-07-01"
          networkInterfaceConfigurations = [
            {
              name = "example"
              properties = {
                dnsSettings = {
                  dnsServers = [
                  ]
                }
                enableAcceleratedNetworking = false
                enableIPForwarding          = false
                ipConfigurations = [
                  {
                    name = "internal"
                    properties = {
                      applicationGatewayBackendAddressPools = [
                      ]
                      applicationSecurityGroups = [
                      ]
                      loadBalancerBackendAddressPools = [
                      ]
                      primary                 = true
                      privateIPAddressVersion = "IPv4"
                      subnet = {
                        id = azapi_resource.subnet.id
                      }
                    }
                  },
                ]
                primary = true
              }
            },
          ]
        }
        osProfile = {
          adminUsername      = "adminuser"
          computerNamePrefix = var.resource_name
          linuxConfiguration = {
            disablePasswordAuthentication = true
            provisionVMAgent              = true
            ssh = {
              publicKeys = [
                {
                  keyData = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC+wWK73dCr+jgQOAxNsHAnNNNMEMWOHYEccp6wJm2gotpr9katuF/ZAdou5AaW1C61slRkHRkpRRX9FA9CYBiitZgvCCz+3nWNN7l/Up54Zps/pHWGZLHNJZRYyAB6j5yVLMVHIHriY49d/GZTZVNB8GoJv9Gakwc/fuEZYYl4YDFiGMBP///TzlI4jhiJzjKnEvqPFki5p2ZRJqcbCiF4pJrxUQR/RXqVFQdbRLZgYfJ8xGB878RENq3yQ39d8dVOkq4edbkzwcUmwwwkYVPIoDGsYLaRHnG+To7FvMeyO7xDVQkMKzopTQV8AuKpyvpqu0a9pWOMaiCyDytO7GGN you@me.com"
                  path    = "/home/adminuser/.ssh/authorized_keys"
                },
              ]
            }
          }
          secrets = [
          ]
        }
        priority = "Regular"
        storageProfile = {
          dataDisks = [
          ]
          imageReference = {
            offer     = "UbuntuServer"
            publisher = "Canonical"
            sku       = "16.04-LTS"
            version   = "latest"
          }
          osDisk = {
            caching      = "ReadWrite"
            createOption = "FromImage"
            managedDisk = {
              storageAccountType = "Standard_LRS"
            }
            osType                  = "Linux"
            writeAcceleratorEnabled = false
          }
        }
      }
    }
    sku = {
      capacity = 1
      name     = "Standard_F2"
      tier     = "Standard"
    }
  }
  schema_validation_enabled = false
  response_export_values    = ["*"]
}

data "azapi_client_config" "current" {}

data "azuread_service_principal" "test" {
  display_name = "Standby Pool Resource Provider"
}

locals {
  vm_contributor_role_id = "/subscriptions/${data.azapi_client_config.current.subscription_id}/providers/Microsoft.Authorization/roleDefinitions/9980e02c-c2be-4d73-94e8-173b1dc7cf3c"
  nw_contributor_role_id = "/subscriptions/${data.azapi_client_config.current.subscription_id}/providers/Microsoft.Authorization/roleDefinitions/4d97b98b-1d4f-4787-a291-c67834d212e7"
  mi_contributor_role_id = "/subscriptions/${data.azapi_client_config.current.subscription_id}/providers/Microsoft.Authorization/roleDefinitions/e40ec5ca-96e0-45a2-b4ff-59039f2c2b59"
}

resource "azapi_resource" "vm-contributor" {
  type      = "Microsoft.Authorization/roleAssignments@2022-04-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = uuidv5("url", "${azapi_resource.resourceGroup.id}/roleAssignments/${data.azuread_service_principal.test.object_id}/9980e02c-c2be-4d73-94e8-173b1dc7cf3c")
  body = {
    properties = {
      principalId      = data.azuread_service_principal.test.object_id
      principalType    = "ServicePrincipal"
      roleDefinitionId = local.vm_contributor_role_id
    }
  }
}

resource "azapi_resource" "nw-contributor" {
  type      = "Microsoft.Authorization/roleAssignments@2022-04-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = uuidv5("url", "${azapi_resource.resourceGroup.id}/roleAssignments/${data.azuread_service_principal.test.object_id}/4d97b98b-1d4f-4787-a291-c67834d212e7")
  body = {
    properties = {
      principalId      = data.azuread_service_principal.test.object_id
      principalType    = "ServicePrincipal"
      roleDefinitionId = local.nw_contributor_role_id
    }
  }
}

resource "azapi_resource" "mi-contributor" {
  type      = "Microsoft.Authorization/roleAssignments@2022-04-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = uuidv5("url", "${azapi_resource.resourceGroup.id}/roleAssignments/${data.azuread_service_principal.test.object_id}/e40ec5ca-96e0-45a2-b4ff-59039f2c2b59")
  body = {
    properties = {
      principalId      = data.azuread_service_principal.test.object_id
      principalType    = "ServicePrincipal"
      roleDefinitionId = local.mi_contributor_role_id
    }
  }
}

resource "azapi_resource" "standbyVirtualMachinePool" {
  type      = "Microsoft.StandbyPool/standbyVirtualMachinePools@2025-03-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = "${var.resource_name}-pool"
  location  = var.location

  body = {
    properties = {
      attachedVirtualMachineScaleSetId = azapi_resource.virtualMachineScaleSet.id
      elasticityProfile = {
        maxReadyCapacity = 2
        minReadyCapacity = 1
      }
      virtualMachineState = "Running"
    }
  }
  schema_validation_enabled = false
  ignore_casing             = false
  ignore_missing_property   = false

  depends_on = [
    azapi_resource.vm-contributor,
    azapi_resource.nw-contributor,
    azapi_resource.mi-contributor,
  ]
}

```



## Arguments Reference

The following arguments are supported:

* `type` - (Required) The type of the resource. This should be set to `Microsoft.StandbyPool/standbyvirtualmachinepools@api-version`. The available api-versions for this resource are: [`2023-12-01-preview`, `2024-03-01`, `2024-03-01-preview`, `2025-03-01`, `2025-10-01`].

* `parent_id` - (Required) The ID of the azure resource in which this resource is created. The allowed values are:  
  `/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}`

* `name` - (Required) Specifies the name of the azure resource. Changing this forces a new resource to be created.

* `body` - (Required) Specifies the configuration of the resource. More information about the arguments in `body` can be found in the [Microsoft documentation](https://learn.microsoft.com/en-us/azure/templates/Microsoft.StandbyPool/standbyvirtualmachinepools?pivots=deployment-language-terraform).

For other arguments, please refer to the [azapi_resource](https://registry.terraform.io/providers/Azure/azapi/latest/docs/resources/resource) documentation.

## Import

 ```shell
 # Azure resource can be imported using the resource id, e.g.
 terraform import azapi_resource.example /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.StandbyPool/standbyvirtualmachinepools/{resourceName}
 
 # It also supports specifying API version by using the resource id with api-version as a query parameter, e.g.
 terraform import azapi_resource.example /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.StandbyPool/standbyvirtualmachinepools/{resourceName}?api-version=2025-10-01
 ```
