# Verify VGN and submodules

Pre-Commit hook that verifies vgn and submodules:

- Each submodule must have a `vgn-version.txt` file at the root that matches vgn version in `go.mod`.
- No submodules can have pending changes.
- Replace is not being used in `go.mod` to redirect `go.kwusen.ca/vgn` to a local directory.

# Install

Add this to the `.pre-commit-config.yaml` file:

``` yaml
- repo: https://github.com/Kwusen/pre-commit-vgn-verify
  rev: 'v0.0.1'
  hooks:
  - id: verify-vgn

```
