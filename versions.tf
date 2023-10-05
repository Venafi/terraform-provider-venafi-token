terraform {
  required_providers {
    random = {
      source = "hashicorp/random"
    }
    venafi-token = {
      source  = "Venafi/venafi-token"
#      version = "0.1.0"
    }
    venafi = {
      source  = "Venafi/venafi"
      version = "0.17.1"
    }
  }
}