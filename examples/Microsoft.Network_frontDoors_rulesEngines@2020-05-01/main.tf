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
