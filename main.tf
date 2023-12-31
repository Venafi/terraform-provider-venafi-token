/*
This is an example Terraform file to show capabilities of the Venafi-token integration.
*/
variable "TPP_URL" {
  default = ""
}

variable "TPP_ZONE" {
  default = ""
}

variable "TRUST_BUNDLE" {
  default = ""
}

resource "venafi-token_credential" "example" {}

resource "random_string" "cn" {
  length  = 5
  special = false
  upper   = false
  numeric = false
}

provider "venafi" {
  alias        = "tpp_token"
  url          = var.TPP_URL
  zone         = var.TPP_ZONE
  trust_bundle = file(var.TRUST_BUNDLE)
  access_token = venafi-token_credential.example.access_token
}

resource "venafi_certificate" "dev_certificate" {
  //Name of the used provider
  provider    = venafi.tpp_token
  common_name = "dev-${random_string.cn.result}.venafi.example.com"

  //Key encryption algorithm
  algorithm = "RSA"

  //DNS aliases
  san_dns = [
    "dev-web01-${random_string.cn.result}.example.com",
    "dev-web02-${random_string.cn.result}.example.com",
  ]

  //IP aliases
  san_ip = [
    "10.1.1.1",
    "192.168.0.1",
  ]

  //Email aliases
  san_email = [
    "dev@venafi.com",
    "dev2@venafi.com",
  ]

  //private key password
  key_password = "newPassw0rd!"
}

//output certificate
output "cert_certificate_dev" {
  value = venafi_certificate.dev_certificate.certificate
}

//output certificate chain
output "cert_chain_dev" {
  value = venafi_certificate.dev_certificate.chain
}

//output private key
output "cert_private_key_dev" {
  sensitive = true
  value = venafi_certificate.dev_certificate.private_key_pem
}