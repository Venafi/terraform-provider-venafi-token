terraform {
  required_providers {
    random = {
      source = "hashicorp/random"
    }
    venafi-token = {
      source = "Venafi/venafi-token"
      version = "99.9.9"
    }
    venafi = {
      source = "Venafi/venafi"
      version = "0.17.0"
    }
  }
}