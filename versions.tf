terraform {
  required_providers {
    random = {
      source = "hashicorp/random"
    }
    venafi-token = {
      source  = "Venafi/venafi-token"
    }
    venafi = {
      source  = "Venafi/venafi"
    }
  }
}