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
  name     = var.resource_name
  location = var.location
}

data "azapi_client_config" "current" {}

resource "azapi_resource" "uai" {
  type      = "Microsoft.ManagedIdentity/userAssignedIdentities@2023-01-31"
  name      = "TestUAI"
  parent_id = azapi_resource.resourceGroup.id
  location  = azapi_resource.resourceGroup.location
  body = {
    properties = {}
  }
  response_export_values = ["properties.principalId"]
}

resource "azapi_resource" "roleAssignments" {
  type      = "Microsoft.Authorization/roleAssignments@2022-04-01"
  name      = "6faae21a-0cd6-4536-8c23-a278823d12ed"
  parent_id = azapi_resource.resourceGroup.id
  body = {
    properties = {
      principalId      = azapi_resource.uai.output.properties.principalId
      principalType    = "ServicePrincipal"
      roleDefinitionId = "/subscriptions/${data.azapi_client_config.current.subscription_id}/providers/Microsoft.Authorization/roleDefinitions/7f951dda-4ed3-4680-a7ca-43fe172d538d"
    }
  }
}
