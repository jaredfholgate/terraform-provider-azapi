---
subcategory: "Microsoft.Storage - Storage"
page_title: "storageAccounts"
description: |-
  Manages a Azure Storage Account.
---

# Microsoft.Storage/storageAccounts - Azure Storage Account

This article demonstrates how to use `azapi` provider to manage the Azure Storage Account resource in Azure.



## Example Usage

### basic

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
  default = "westeurope"
}

resource "azapi_resource" "resourceGroup" {
  type     = "Microsoft.Resources/resourceGroups@2020-06-01"
  name     = var.resource_name
  location = var.location
}

resource "azapi_resource" "storageAccount" {
  type      = "Microsoft.Storage/storageAccounts@2021-09-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = var.resource_name
  location  = var.location
  body = {
    kind = "StorageV2"
    properties = {
      accessTier                   = "Hot"
      allowBlobPublicAccess        = true
      allowCrossTenantReplication  = true
      allowSharedKeyAccess         = true
      defaultToOAuthAuthentication = false
      encryption = {
        keySource = "Microsoft.Storage"
        services = {
          queue = {
            keyType = "Service"
          }
          table = {
            keyType = "Service"
          }
        }
      }
      isHnsEnabled      = false
      isNfsV3Enabled    = false
      isSftpEnabled     = false
      minimumTlsVersion = "TLS1_2"
      networkAcls = {
        defaultAction = "Allow"
      }
      publicNetworkAccess      = "Enabled"
      supportsHttpsTrafficOnly = true
    }
    sku = {
      name = "Standard_LRS"
    }
  }
  schema_validation_enabled = false
  response_export_values    = ["*"]
}


```

### with_private_endpoint

```hcl
terraform {
  required_providers {
    azapi = {
      source = "Azure/azapi"
    }
  }
}

provider "azapi" {
  subscription_id            = var.subscription_id
  skip_provider_registration = false
}

variable "resource_name" {
  type    = string
  default = "acctest0001"
}

variable "location" {
  type    = string
  default = "westeurope"
}

variable "subscription_id" {
  type    = string
  default = "00000000-0000-0000-0000-000000000000"
}

variable "vm_admin_username" {
  type    = string
  default = "adminuser"
}

variable "vm_admin_password" {
  type        = string
  description = "The administrator password for the virtual machine"
  sensitive   = true
}

data "azapi_client_config" "current" {}

resource "azapi_resource" "resourceGroup" {
  type     = "Microsoft.Resources/resourceGroups@2020-06-01"
  name     = var.resource_name
  location = var.location
}

resource "azapi_resource" "virtualNetwork" {
  type      = "Microsoft.Network/virtualNetworks@2023-11-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = "${var.resource_name}-vnet"
  location  = var.location
  body = {
    properties = {
      addressSpace = {
        addressPrefixes = ["10.0.0.0/16"]
      }
    }
  }
  schema_validation_enabled = false
  lifecycle {
    ignore_changes = [body.properties.subnets]
  }
}

resource "azapi_resource" "subnetMain" {
  type      = "Microsoft.Network/virtualNetworks/subnets@2023-11-01"
  parent_id = azapi_resource.virtualNetwork.id
  name      = "main"
  body = {
    properties = {
      addressPrefix = "10.0.1.0/24"
    }
  }
  schema_validation_enabled = false
}

resource "azapi_resource" "subnetVm" {
  type      = "Microsoft.Network/virtualNetworks/subnets@2023-11-01"
  parent_id = azapi_resource.virtualNetwork.id
  name      = "vm"
  body = {
    properties = {
      addressPrefix = "10.0.2.0/24"
    }
  }
  schema_validation_enabled = false
}

resource "azapi_resource" "subnetBastion" {
  type      = "Microsoft.Network/virtualNetworks/subnets@2023-11-01"
  parent_id = azapi_resource.virtualNetwork.id
  name      = "AzureBastionSubnet"
  body = {
    properties = {
      addressPrefix = "10.0.3.0/24"
    }
  }
  schema_validation_enabled = false
}

resource "azapi_resource" "privateDnsZoneBlob" {
  type      = "Microsoft.Network/privateDnsZones@2020-06-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = "privatelink.blob.core.windows.net"
  location  = "global"
  body = {
    properties = {}
  }
}

resource "azapi_resource" "privateDnsZoneQueue" {
  type      = "Microsoft.Network/privateDnsZones@2020-06-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = "privatelink.queue.core.windows.net"
  location  = "global"
  body = {
    properties = {}
  }
}

resource "azapi_resource" "privateDnsZoneWeb" {
  type      = "Microsoft.Network/privateDnsZones@2020-06-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = "privatelink.web.core.windows.net"
  location  = "global"
  body = {
    properties = {}
  }
}

resource "azapi_resource" "privateDnsZoneVnetLinkBlob" {
  type      = "Microsoft.Network/privateDnsZones/virtualNetworkLinks@2020-06-01"
  parent_id = azapi_resource.privateDnsZoneBlob.id
  name      = "blob"
  location  = "global"
  body = {
    properties = {
      registrationEnabled = false
      virtualNetwork = {
        id = azapi_resource.virtualNetwork.id
      }
    }
  }
}

resource "azapi_resource" "privateDnsZoneVnetLinkQueue" {
  type      = "Microsoft.Network/privateDnsZones/virtualNetworkLinks@2020-06-01"
  parent_id = azapi_resource.privateDnsZoneQueue.id
  name      = "queue"
  location  = "global"
  body = {
    properties = {
      registrationEnabled = false
      virtualNetwork = {
        id = azapi_resource.virtualNetwork.id
      }
    }
  }
}

resource "azapi_resource" "privateDnsZoneVnetLinkWeb" {
  type      = "Microsoft.Network/privateDnsZones/virtualNetworkLinks@2020-06-01"
  parent_id = azapi_resource.privateDnsZoneWeb.id
  name      = "web"
  location  = "global"
  body = {
    properties = {
      registrationEnabled = false
      virtualNetwork = {
        id = azapi_resource.virtualNetwork.id
      }
    }
  }
}

resource "azapi_resource" "publicIp" {
  type      = "Microsoft.Network/publicIPAddresses@2023-11-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = "${var.resource_name}-ip"
  location  = var.location
  body = {
    sku = {
      name = "Standard"
    }
    properties = {
      publicIPAllocationMethod = "Static"
    }
  }
}

resource "azapi_resource" "bastionHost" {
  type      = "Microsoft.Network/bastionHosts@2023-11-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = "${var.resource_name}-bastion"
  location  = var.location
  body = {
    sku = {
      name = "Basic"
    }
    properties = {
      ipConfigurations = [
        {
          name = "configuration"
          properties = {
            subnet = {
              id = azapi_resource.subnetBastion.id
            }
            publicIPAddress = {
              id = azapi_resource.publicIp.id
            }
          }
        }
      ]
    }
  }
}

resource "azapi_resource" "networkInterface" {
  type      = "Microsoft.Network/networkInterfaces@2023-11-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = "${var.resource_name}-nic"
  location  = var.location
  body = {
    properties = {
      ipConfigurations = [
        {
          name = "internal"
          properties = {
            subnet = {
              id = azapi_resource.subnetVm.id
            }
            privateIPAllocationMethod = "Dynamic"
          }
        }
      ]
    }
  }
  response_export_values = ["identity.principalId"]
}

resource "azapi_resource" "windowsVirtualMachine" {
  type      = "Microsoft.Compute/virtualMachines@2023-03-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = var.resource_name
  location  = var.location
  identity {
    type = "SystemAssigned"
  }
  body = {
    properties = {
      hardwareProfile = {
        vmSize = "Standard_F2"
      }
      osProfile = {
        computerName  = var.resource_name
        adminUsername = var.vm_admin_username
        adminPassword = var.vm_admin_password
        windowsConfiguration = {
          provisionVMAgent             = true
          enableVMAgentPlatformUpdates = true
        }
      }
      networkProfile = {
        networkInterfaces = [
          {
            id = azapi_resource.networkInterface.id
            properties = {
              primary = true
            }
          }
        ]
      }
      storageProfile = {
        imageReference = {
          publisher = "MicrosoftWindowsServer"
          offer     = "WindowsServer"
          sku       = "2022-Datacenter"
          version   = "latest"
        }
        osDisk = {
          createOption = "FromImage"
          caching      = "ReadWrite"
          managedDisk = {
            storageAccountType = "Standard_LRS"
          }
        }
      }
    }
  }
  schema_validation_enabled = false
  response_export_values    = ["identity.principalId"]
}

resource "azapi_resource" "storageAccount" {
  type      = "Microsoft.Storage/storageAccounts@2023-05-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = "${var.resource_name}sa"
  location  = var.location
  body = {
    kind = "StorageV2"
    properties = {
      accessTier                   = "Hot"
      allowBlobPublicAccess        = false
      allowCrossTenantReplication  = false
      allowSharedKeyAccess         = true
      defaultToOAuthAuthentication = false
      dnsEndpointType              = "Standard"
      encryption = {
        keySource = "Microsoft.Storage"
        services = {
          blob = {
            enabled = true
            keyType = "Account"
          }
          file = {
            enabled = true
            keyType = "Account"
          }
        }
      }
      isHnsEnabled       = false
      isLocalUserEnabled = true
      isNfsV3Enabled     = false
      isSftpEnabled      = false
      minimumTlsVersion  = "TLS1_2"
      networkAcls = {
        bypass              = "AzureServices"
        defaultAction       = "Allow"
        ipRules             = []
        resourceAccessRules = []
        virtualNetworkRules = []
      }
      publicNetworkAccess      = "Disabled"
      supportsHttpsTrafficOnly = true
    }
    sku = {
      name = "Standard_LRS"
    }
  }
}

resource "azapi_resource" "roleAssignment" {
  type      = "Microsoft.Authorization/roleAssignments@2022-04-01"
  parent_id = azapi_resource.storageAccount.id
  name      = uuidv5("url", "${azapi_resource.storageAccount.id}/roleAssignments/${azapi_resource.windowsVirtualMachine.output.identity.principalId}")
  body = {
    properties = {
      principalId      = azapi_resource.windowsVirtualMachine.output.identity.principalId
      principalType    = "ServicePrincipal"
      roleDefinitionId = "/subscriptions/${data.azapi_client_config.current.subscription_id}/providers/Microsoft.Authorization/roleDefinitions/b24988ac-6180-42a0-ab88-20f7382dd24c"
    }
  }
}

resource "azapi_resource" "privateEndpointBlob" {
  type      = "Microsoft.Network/privateEndpoints@2023-11-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = "${var.resource_name}-private-endpoint-blob"
  location  = var.location
  body = {
    properties = {
      subnet = {
        id = azapi_resource.subnetMain.id
      }
      privateLinkServiceConnections = [
        {
          name = "${var.resource_name}-private-endpoint-connection"
          properties = {
            privateLinkServiceId = azapi_resource.storageAccount.id
            groupIds             = ["blob"]
          }
        }
      ]
    }
  }
  schema_validation_enabled = false
}

resource "azapi_resource" "privateEndpointBlobDnsZoneGroup" {
  type      = "Microsoft.Network/privateEndpoints/privateDnsZoneGroups@2023-11-01"
  parent_id = azapi_resource.privateEndpointBlob.id
  name      = "storage-private-endpoint-dns-zone-group"
  body = {
    properties = {
      privateDnsZoneConfigs = [
        {
          name = "privatelink-blob-core-windows-net"
          properties = {
            privateDnsZoneId = azapi_resource.privateDnsZoneBlob.id
          }
        }
      ]
    }
  }
}

resource "azapi_resource" "privateEndpointQueue" {
  type      = "Microsoft.Network/privateEndpoints@2023-11-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = "${var.resource_name}-private-endpoint-queue"
  location  = var.location
  body = {
    properties = {
      subnet = {
        id = azapi_resource.subnetMain.id
      }
      privateLinkServiceConnections = [
        {
          name = "${var.resource_name}-private-endpoint-connection-q"
          properties = {
            privateLinkServiceId = azapi_resource.storageAccount.id
            groupIds             = ["queue"]
          }
        }
      ]
    }
  }
  schema_validation_enabled = false
}

resource "azapi_resource" "privateEndpointQueueDnsZoneGroup" {
  type      = "Microsoft.Network/privateEndpoints/privateDnsZoneGroups@2023-11-01"
  parent_id = azapi_resource.privateEndpointQueue.id
  name      = "storage-private-endpoint-dns-zone-group-q"
  body = {
    properties = {
      privateDnsZoneConfigs = [
        {
          name = "privatelink-queue-core-windows-net"
          properties = {
            privateDnsZoneId = azapi_resource.privateDnsZoneQueue.id
          }
        }
      ]
    }
  }
}

resource "azapi_resource" "privateEndpointWeb" {
  type      = "Microsoft.Network/privateEndpoints@2023-11-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = "${var.resource_name}-private-endpoint-web"
  location  = var.location
  body = {
    properties = {
      subnet = {
        id = azapi_resource.subnetMain.id
      }
      privateLinkServiceConnections = [
        {
          name = "${var.resource_name}-private-endpoint-connection-web"
          properties = {
            privateLinkServiceId = azapi_resource.storageAccount.id
            groupIds             = ["web"]
          }
        }
      ]
    }
  }
  schema_validation_enabled = false
}

resource "azapi_resource" "privateEndpointWebDnsZoneGroup" {
  type      = "Microsoft.Network/privateEndpoints/privateDnsZoneGroups@2023-11-01"
  parent_id = azapi_resource.privateEndpointWeb.id
  name      = "storage-private-endpoint-dns-zone-group-web"
  body = {
    properties = {
      privateDnsZoneConfigs = [
        {
          name = "privatelink-web-core-windows-net"
          properties = {
            privateDnsZoneId = azapi_resource.privateDnsZoneWeb.id
          }
        }
      ]
    }
  }
}

```



## Arguments Reference

The following arguments are supported:

* `type` - (Required) The type of the resource. This should be set to `Microsoft.Storage/storageAccounts@api-version`. The available api-versions for this resource are: [`2015-05-01-preview`, `2015-06-15`, `2016-01-01`, `2016-05-01`, `2016-12-01`, `2017-06-01`, `2017-10-01`, `2018-02-01`, `2018-03-01-preview`, `2018-07-01`, `2018-11-01`, `2019-04-01`, `2019-06-01`, `2020-08-01-preview`, `2021-01-01`, `2021-02-01`, `2021-04-01`, `2021-06-01`, `2021-08-01`, `2021-09-01`, `2022-05-01`, `2022-09-01`, `2023-01-01`, `2023-04-01`, `2023-05-01`, `2024-01-01`, `2025-01-01`, `2025-06-01`].

* `parent_id` - (Required) The ID of the azure resource in which this resource is created. The allowed values are:  
  `/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}`

* `name` - (Required) Specifies the name of the azure resource. Changing this forces a new resource to be created.

* `body` - (Required) Specifies the configuration of the resource. More information about the arguments in `body` can be found in the [Microsoft documentation](https://learn.microsoft.com/en-us/azure/templates/Microsoft.Storage/storageAccounts?pivots=deployment-language-terraform).

For other arguments, please refer to the [azapi_resource](https://registry.terraform.io/providers/Azure/azapi/latest/docs/resources/resource) documentation.

## Import

 ```shell
 # Azure resource can be imported using the resource id, e.g.
 terraform import azapi_resource.example /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Storage/storageAccounts/{resourceName}
 
 # It also supports specifying API version by using the resource id with api-version as a query parameter, e.g.
 terraform import azapi_resource.example /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Storage/storageAccounts/{resourceName}?api-version=2025-06-01
 ```
