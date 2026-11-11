## 0.2.2 (November 11th, 2026)
- Builds with Go 1.27.3. Earlier Go versions accept certificate chains from the compromised eMudhra intermediate CAs; Go 1.27.3 rejects them, so the provider no longer trusts TLS servers presenting those chains.
- Updates the VCert SDK to v5.14.0 ([#53](https://github.com/Venafi/terraform-provider-venafi-token/pull/53)).
- Security: credentials are redacted from import logs ([#36](https://github.com/Venafi/terraform-provider-venafi-token/pull/36)), and the PKCS#12 client certificate attribute is marked as sensitive ([#50](https://github.com/Venafi/terraform-provider-venafi-token/pull/50)).
- Replaces unmaintained dependencies and updates dependency libraries ([#39](https://github.com/Venafi/terraform-provider-venafi-token/pull/39), [#45](https://github.com/Venafi/terraform-provider-venafi-token/pull/45), [#47](https://github.com/Venafi/terraform-provider-venafi-token/pull/47), [#49](https://github.com/Venafi/terraform-provider-venafi-token/pull/49)).

## 0.2.1 (January 23rd, 2025)
- Updates dependency libraries, including gRPC, to fix vulnerabilities.

## 0.2.0 (October 14th, 2023)
- Upgrades the Terraform plugin framework to v1.4.1 to support Terraform v1.6.1.
- Fixed crash issue when no trust bundle was passed in the credential resource import string. ([#12](https://github.com/Venafi/terraform-provider-venafi-token/issues/12))

## 0.1.0 (Occtober 1, 2023)
- Initial release
- Added ability to refresh TLSPDC access tokens
- Added three authentication methods: refresh token, client certificate and username/password (deprecated)