# Release Signing

Planwright's release integrity model uses a human-controlled OpenPGP key to sign checksum manifests.
- [Trust Root](#trust-root)
- [Maintainer Rules](#maintainer-rules)
- [User Verification](#user-verification)
- [Provenance Attestations](#provenance-attestations)

Current signing-key status:
- the planned Planwright release signing key fingerprint is `6DB2 3F44 3178 8F6B 49A7 A3E4 E87B 0A3D CF5E FCA9`
- the first signed release must use this fingerprint unless the key is rotated before tagging
- signatures should not be treated as trusted unless the public key fingerprint is verified through this file or another maintainer-controlled channel

# Trust Root

The trust root is the release key fingerprint published here after the key is created.

Field | Value
:--- | :---
Release key fingerprint | `6DB2 3F44 3178 8F6B 49A7 A3E4 E87B 0A3D CF5E FCA9`
First planned signed release | `v0.11.0`
Signature scope | `SHA2-256SUMS` and `SHA2-512SUMS`

The `public.key` asset on a GitHub release is a convenience copy of the public key. It helps users import the key but it does not replace fingerprint verification.

# Maintainer Rules

Maintainers must keep the release key boundary explicit:
- never commit private key material
- never paste private key material into issues, pull requests, logs or release notes
- store release signing secrets only in the protected GitHub `release` environment
- require maintainer approval before release workflow access to signing secrets
- rotate the key if private key material or the passphrase is exposed
- do not silently replace the key; add a dated note here and keep the previous fingerprint for old releases

The release workflow expects these GitHub Actions secrets:
- `RELEASE_SIGNING_PRIVATE_KEY`
- `RELEASE_SIGNING_PASSPHRASE`

The release workflow expects this GitHub Actions variable:
- `RELEASE_SIGNING_KEY_FINGERPRINT`

# User Verification

After downloading a binary and the release integrity assets, import the release public key:
```bash
gpg --import ./public.key
```

Compare the imported fingerprint with the fingerprint documented above:
```bash
gpg --fingerprint
```

Verify the checksum manifest signature:
```bash
gpg --verify ./SHA2-256SUMS.sig ./SHA2-256SUMS
```

Verify the downloaded binary against the signed manifest:
```bash
sha256sum -c ./SHA2-256SUMS --ignore-missing
```

The signed manifests also cover release SBOM files when they are attached to the release.

# Provenance Attestations

Planwright release assets are also covered by GitHub artefact attestations created by the release workflow.

Verify the downloaded Linux binary attestation:
```bash
gh attestation verify ./planwright_linux_amd64 -R steadytao/planwright
```

Verify the downloaded Windows binary attestation:
```bash
gh attestation verify ./planwright_windows_amd64.exe -R steadytao/planwright
```

The attestation path proves workflow-backed provenance for the file digest. It complements the human-controlled OpenPGP checksum signature but does not replace fingerprint verification of the release signing key.
