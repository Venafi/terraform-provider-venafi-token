terraform {
  required_providers {
    random = {
      source = "hashicorp/random"
    }
    venafi = {
      source = "Venafi/venafi-token"
    }
  }
  required_version = ">= 0.13"
}
