# cert-manager Vault Issuer (WIP)

> ⚠️ **Work in Progress**  
> This repository is under active development. It is not yet ready for production use and may change significantly before a stable release.

## Overview

This project provides an out-of-tree [cert-manager](https://github.com/cert-manager/cert-manager) issuer for integrating with Vault-based secret management systems.

It is a continuation of the [in-tree Vault issuer](https://cert-manager.io/docs/configuration/vault/) previously maintained within cert-manager itself.  
By extracting it into a dedicated repository, we aim to:
- Improve maintainability and release cadence
- Enable independent versioning and feature development
- Allow faster iteration and community contributions

## Status

The code and documentation are being refactored and migrated from cert-manager’s core repository.  
Once complete, this issuer will become the canonical Vault integration for cert-manager.

## Trademark Notice

This project is an independent integration for cert-manager.  
It is **not affiliated with, endorsed by, or sponsored by HashiCorp or IBM**.  
“HashiCorp” and “Vault” are trademarks of HashiCorp, Inc., and are used here only to describe compatibility with the Vault product.

## Contributing

Contributions are welcome!  
Until the migration is complete, please refer to ongoing discussions and design proposals in the [cert-manager repository](https://github.com/cert-manager/cert-manager/issues).

## License

This project follows the same licensing model as cert-manager.  
See [LICENSE](./LICENSE) for details.
