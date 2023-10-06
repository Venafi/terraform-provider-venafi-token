[![Venafi](https://raw.githubusercontent.com/Venafi/.github/master/images/Venafi_logo.png)](https://www.venafi.com/)
[![MPL 2.0 License](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](https://opensource.org/licenses/MPL-2.0)
![Community Supported](https://img.shields.io/badge/Support%20Level-Community-brightgreen)
![Compatible with TPP 17.3+ & VaaS](https://img.shields.io/badge/Compatibility-TPP%2017.3+%20%26%20VaaS-f9a90c) 
_**This open source project is community-supported.** To report a problem or share an idea, use
**[Issues](../../issues)**; and if you have a suggestion for fixing the issue, please include those details, too.
In addition, use **[Pull Requests](../../pulls)** to contribute actual bug fixes or proposed enhancements.
We welcome and appreciate all contributions. Got questions or want to discuss something with our team?
**[Join us on Slack](https://join.slack.com/t/venafi-integrations/shared_invite/zt-i8fwc379-kDJlmzU8OiIQOJFSwiA~dg)**!_

# Venafi-token Provider for HashiCorp Terraform

The `venafi-token` provider streamlines the process of rotating tokens from a Trust Protection Platform (TLSPDC) 
instance. It provides a resource that allows tokens to be managed and rotated when they expire or a custom refresh 
threshold (in days) is met.

## Requirements

### Protection of the terraform state file

Make sure that you are protecting your terraform state file as per the best practices by Hashicorp: [https://developer.hashicorp.com/terraform/language/state/sensitive-data](https://developer.hashicorp.com/terraform/language/state/sensitive-data).  
This is an important step to prevent data breaches or leaks of sensitive data like usernames, passwords, tokens, secrets, etc.

### Trust between Terraform and Trust Protection Platform

The Trust Protection Platform REST API (WebSDK) must be secured with a
certificate. Generally, the certificate is issued by a CA that is not publicly
trusted so establishing trust is a critical part of your setup.

Two methods can be used to establish trust. Both require the trust anchor
(root CA certificate) of the WebSDK certificate. If you have administrative
access, you can import the root certificate into the trust store for your
operating system. If you don't have administrative access, or prefer not to
make changes to your system configuration, save the root certificate to a file
in PEM format (e.g. /opt/venafi/bundle.pem) and include it using the
`trust_bundle` parameter of your Venafi provider.

## Setup

The `venafi-token` Provider for HashiCorp Terraform is an officially verified integration. As such, releases are 
published to the [Terraform Registry](https://registry.terraform.io/providers/Venafi/venafi/latest) where they are 
available for `terraform init` to automatically download whenever the provider is referenced by a configuration file.  
No setup steps are required to use an official release of this provider other than to download and install Terraform 
itself.

To use a pre-release or custom built version of this provider, manually install the plugin binary into
[required directory](https://www.terraform.io/docs/commands/init.html#plugin-installation) using the prescribed
[subdirectory structure](https://www.terraform.io/docs/configuration/provider-requirements.html#source-addresses)
that must align with how the provider is referenced in the `required_providers` block of the configuration file.


## Usage

1. Declare the `venafi-token` provider as required for your plan:
   ```terraform
   terraform {
     required_providers {
       venafi-token = {
         source  = "Venafi/venafi-token"
       }
     }
     required_version = ">= 0.13"
   }
   ```
2. Declare a credential resource in your terraform plan:
   ```terraform
   resource "venafi-token_credential" "example" {}
   ```

3. Import the credential resource.  
!> NOTE: It is very important that the resource is imported and not created. It holds sensitive data should not be displayed 
during the terraform plan.  
!> [Detailed documentation on how to build the import string can be found here.](https://github.com/Venafi/terraform-provider-venafi-token/blob/main/docs/resources/credential.md)
   ```sh
   terraform import venafi-token_credential.example 'url=<value>,trust_bundle=<value>,refresh_token=<value>'
   ```
4. Assign the `access token` from `venafi-token_credential.example` to the venafi provider block:
   ```terraform
   provider "venafi" {
     alias        = "tpp_token"
     url          = var.TPP_URL
     zone         = var.TPP_ZONE
     trust_bundle = file(var.TRUST_BUNDLE)
     access_token = venafi-token_credential.example.access_token
   }
   ```
5. Run your terraform plan:
   ```shell
   terraform apply
   ```
   
## License

Copyright &copy; Venafi, Inc. All rights reserved.

This solution is licensed under the Mozilla Public License, Version 2.0. See [LICENSE](./LICENSE) for the full license text.

Please direct questions/comments to opensource@venafi.com.
