---
subcategory: "Microsoft.Network - Various Networking Services"
page_title: "frontDoors/rulesEngines"
description: |-
  Manages a Azure Front Door (classic) Rules Engine configuration and rules.
---

# Microsoft.Network/frontDoors/rulesEngines - Azure Front Door (classic) Rules Engine configuration and rules

This article demonstrates how to use `azapi` provider to manage the Azure Front Door (classic) Rules Engine configuration and rules resource in Azure.



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
  default = "westeurope"
}

resource "azapi_resource" "resourceGroup" {
  type     = "Microsoft.Resources/resourceGroups@2020-06-01"
  name     = var.resource_name
  location = var.location
}

locals {
  backend_name        = "backend-bing"
  endpoint_name       = "frontend-endpoint"
  health_probe_name   = "health-probe"
  load_balancing_name = "load-balancing-setting"
}

resource "azapi_resource" "frontDoor" {
  type      = "Microsoft.Network/frontDoors@2020-05-01"
  parent_id = azapi_resource.resourceGroup.id
  name      = "acctest-FD-test"
  location  = "Global"
  body = {
    properties = {
      enabledState = "Enabled"
      backendPoolsSettings = {
        enforceCertificateNameCheck = "Disabled"
      }
      backendPools = [
        {
          name = local.backend_name
          properties = {
            backends = [
              {
                address           = "www.bing.com"
                backendHostHeader = "www.bing.com"
                httpPort          = 80
                httpsPort         = 443
                priority          = 1
                weight            = 50
                enabledState      = "Enabled"
              }
            ]
            loadBalancingSettings = {
              id = "${azapi_resource.resourceGroup.id}/providers/Microsoft.Network/frontDoors/acctest-FD-test/loadBalancingSettings/${local.load_balancing_name}"
            }
            healthProbeSettings = {
              id = "${azapi_resource.resourceGroup.id}/providers/Microsoft.Network/frontDoors/acctest-FD-test/healthProbeSettings/${local.health_probe_name}"
            }
          }
        }
      ]
      loadBalancingSettings = [
        {
          name = local.load_balancing_name
          properties = {
            sampleSize                    = 4
            successfulSamplesRequired     = 2
            additionalLatencyMilliseconds = 0
          }
        }
      ]
      healthProbeSettings = [
        {
          name = local.health_probe_name
          properties = {
            path              = "/"
            protocol          = "Http"
            intervalInSeconds = 120
            healthProbeMethod = "HEAD"
            enabledState      = "Enabled"
          }
        }
      ]
      frontendEndpoints = [
        {
          name = local.endpoint_name
          properties = {
            hostName = "acctest-FD-test.azurefd.net"
          }
        }
      ]
      routingRules = [
        {
          name = "routing-rule"
          properties = {
            acceptedProtocols = ["Http", "Https"]
            patternsToMatch   = ["/*"]
            enabledState      = "Enabled"
            frontendEndpoints = [
              {
                id = "${azapi_resource.resourceGroup.id}/providers/Microsoft.Network/frontDoors/acctest-FD-test/frontendEndpoints/${local.endpoint_name}"
              }
            ]
            routeConfiguration = {
              "@odata.type"      = "#Microsoft.Azure.FrontDoor.Models.FrontdoorForwardingConfiguration"
              forwardingProtocol = "MatchRequest"
              backendPool = {
                id = "${azapi_resource.resourceGroup.id}/providers/Microsoft.Network/frontDoors/acctest-FD-test/backendPools/${local.backend_name}"
              }
            }
          }
        }
      ]
    }
  }
  schema_validation_enabled = false
  response_export_values    = ["*"]
}

resource "azapi_resource" "rulesEngine" {
  type      = "Microsoft.Network/frontDoors/rulesEngines@2020-05-01"
  parent_id = azapi_resource.frontDoor.id
  name      = var.resource_name
  body = {
    properties = {
      rules = [
        {
          name     = var.resource_name
          priority = 0
          action = {
            routeConfigurationOverride = {
              redirectType     = "Found"
              redirectProtocol = "HttpsOnly"
              customHost       = "customhost.org"
              "@odata.type"    = "#Microsoft.Azure.FrontDoor.Models.FrontdoorRedirectConfiguration"
            }
          }
          matchProcessingBehavior = "Continue"
        }
      ]
    }
  }
}

```



## Arguments Reference

The following arguments are supported:

* `type` - (Required) The type of the resource. This should be set to `Microsoft.Network/frontDoors/rulesEngines@api-version`. The available api-versions for this resource are: [`2020-01-01`, `2020-04-01`, `2020-05-01`, `2021-06-01`, `2025-10-01`].

* `parent_id` - (Required) The ID of the azure resource in which this resource is created. The allowed values are:  
  `/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/frontDoors/{resourceName}`

* `name` - (Required) Specifies the name of the azure resource. Changing this forces a new resource to be created.

* `body` - (Required) Specifies the configuration of the resource. More information about the arguments in `body` can be found in the [Microsoft documentation](https://learn.microsoft.com/en-us/azure/templates/Microsoft.Network/frontDoors/rulesEngines?pivots=deployment-language-terraform).

For other arguments, please refer to the [azapi_resource](https://registry.terraform.io/providers/Azure/azapi/latest/docs/resources/resource) documentation.

## Import

 ```shell
 # Azure resource can be imported using the resource id, e.g.
 terraform import azapi_resource.example /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/frontDoors/{resourceName}/rulesEngines/{resourceName}
 
 # It also supports specifying API version by using the resource id with api-version as a query parameter, e.g.
 terraform import azapi_resource.example /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/frontDoors/{resourceName}/rulesEngines/{resourceName}?api-version=2025-10-01
 ```
