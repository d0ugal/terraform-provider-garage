# Publishing to Terraform Registry

This guide explains how to publish the `terraform-provider-garage` to the Terraform Registry.

## Prerequisites

1. **GPG Key Pair**: You need a GPG key to sign releases
2. **GitHub Secrets**: Store GPG key and passphrase as secrets
3. **Terraform Registry Account**: Sign in with your GitHub account

## Step 1: Generate GPG Key

If you don't have a GPG key, generate one:

```bash
gpg --full-generate-key
# Choose RSA and RSA, 4096 bits, no expiration
# Save the passphrase securely!
```

Export your private key:

```bash
gpg --armor --export-secret-keys YOUR_KEY_ID > gpg-private-key.asc
```

Get your key fingerprint:

```bash
gpg --list-secret-keys --keyid-format LONG
# Look for the line starting with "sec" and note the key ID after the "/"
```

## Step 2: Add GitHub Secrets

Go to your repository settings → Secrets and variables → Actions, and add:

- **`GPG_PRIVATE_KEY`**: The contents of `gpg-private-key.asc` (the entire ASCII-armored key)
- **`GPG_PASSPHRASE`**: The passphrase you set when creating the GPG key
- **`RELEASE_TOKEN`**: Already set via Terraform (personal access token with `repo` scope)

## Step 3: Create Documentation (Optional but Recommended)

The Terraform Registry requires documentation. Create:

- `docs/index.md` - Provider overview
- `docs/resources/garage_key.md` - Key resource documentation
- `docs/resources/garage_bucket.md` - Bucket resource documentation
- `docs/resources/garage_bucket_key.md` - Bucket key resource documentation

## Step 4: Release Process

The release process is automated via GitHub Actions:

1. **Release Please** creates a release PR when you push commits with conventional commit messages
2. When the PR is merged, Release Please creates a GitHub release with a tag (e.g., `v0.0.3`)
3. The **Release workflow** (`.github/workflows/release.yml`) automatically:
   - Builds binaries for all supported platforms (Linux, macOS, Windows, FreeBSD)
   - Creates SHA256 checksums
   - Signs the checksums with GPG
   - Uploads everything to the GitHub release

## Step 5: Publish to Terraform Registry

Once a release is created on GitHub:

1. Go to https://registry.terraform.io/publish
2. Sign in with your GitHub account
3. Click "Publish Provider"
4. Select `d0ugal/terraform-provider-garage`
5. Follow the prompts to publish

The registry will automatically detect new releases and make them available.

## Verification

After publishing, verify the provider is available:

```bash
terraform init
# Should download from registry.terraform.io/d0ugal/garage
```

## Troubleshooting

- **Release workflow fails**: Check that `GPG_PRIVATE_KEY` and `GPG_PASSPHRASE` secrets are set correctly
- **Registry doesn't detect release**: Ensure the release includes:
  - Binaries for multiple platforms
  - `terraform-provider-garage_X.X.X_SHA256SUMS` file
  - `terraform-provider-garage_X.X.X_SHA256SUMS.sig` file
  - `terraform-registry-manifest.json` file
- **Provider not found**: Wait a few minutes after publishing - the registry needs time to index


