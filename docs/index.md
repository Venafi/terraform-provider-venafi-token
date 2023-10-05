---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "venafi-token Provider"
subcategory: ""
description: |-
  This is for refreshing Venafi tokens for use with venafi-provider.
---

# venafi-token Provider

[Venafi](https://www.venafi.com) is the enterprise platform for Machine Identity Protection. The `venafi-token` 
provider streamlines the process of rotating tokens from a Trust Protection Platform (TLSPDC) instance. It provides 
resources that allow tokens to be managed and rotated when they expire or a custom refresh threshold (in days) is met. 

!> NOTE: The resource cannot be created from a terraform plan. It must be imported using `terraform import`.

## Example Usage 

### Refresh Token

```terraform
required_providers {
  venafi-token = {
    source  = "Venafi/venafi-token"
    version = "~> 0.1.0"
  }
  venafi = {
    source  = "Venafi/venafi"
    version = "~> 0.17.0"
  }
}

resource "venafi-token_credential" "example" {
  # This resource MUST be imported.
  # Retrieve a new refresh token using /vedauth/authorize/oauth endpoint of your TPP instance. 
  # Copy the "refresh_token".
  # Generate an import string that includes the `refresh_token` attribute. 
  # See `venafi-token_credential` resource page for import string details.
  # Then run `terraform import venafi-token_credential.example 'refresh_token=xxx,url=...'`
}

provider "venafi" {
  url          = "https://my.tpp.instance.com"
  zone         = "Integrations\\terraform"
  trust_bundle = file("/path/to/my/bundle.cer")
  access_token = venafi-token_credential.example.access_token
}
```

### Client Certificate

```terraform
required_providers {
  venafi-token = {
    source  = "Venafi/venafi-token"
    version = "~> 0.1.0"
  }
  venafi = {
    source  = "Venafi/venafi"
    version = "~> 0.17.0"
  }
}

resource "venafi-token_credential" "example" {
  # This resource MUST be imported.
  # Generate an import string that includes `p12_cert_filename` and `p12_cert_password` attributes. 
  # See `venafi-token_credential` resource page for import string details.
  # Then run `terraform import venafi-token_credential.example 'p12_cert_filename=/path/to/file,p12_cert_password=...'`
}

provider "venafi" {
  url          = "https://my.tpp.instance.com"
  zone         = "Integrations\\terraform"
  trust_bundle = file("/path/to/my/bundle.cer")
  access_token = venafi-token_credential.example.access_token
}
```

### Username and Password

```terraform
required_providers {
  venafi-token = {
    source  = "Venafi/venafi-token"
    version = "~> 0.1.0"
  }
  venafi = {
    source  = "Venafi/venafi"
    version = "~> 0.17.0"
  }
}

resource "venafi-token_credential" "example" {
  # This resource MUST be imported.
  # Generate an import string that includes `username` and `password` attributes. 
  # See `venafi-token_credential` resource page for import string details.
  # Then run `terraform import venafi-token_credential.example 'username=user123,password=x123...'`
}

provider "venafi" {
  url          = "https://my.tpp.instance.com"
  zone         = "Integrations\\terraform"
  trust_bundle = file("/path/to/my/bundle.cer")
  access_token = venafi-token_credential.example.access_token
}
```

<!-- schema generated by tfplugindocs -->
## Schema