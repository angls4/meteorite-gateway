# meteorite gateway

This is the meteorite fork of [QuantumNous/new-api](https://github.com/QuantumNous/new-api), the gateway of meteorite.run (stage 1).

- Based on release `v1.0.0-rc.51` (git remote `upstream`). Record the new base here whenever the fork is rebased.
- License: AGPL-3.0, the same as new-api. The full text is in `LICENSE`; the running web UI links to this repository from its footer, so users can get the source.
- Settings that do not need code (name, logo, footer, sign-up options) are applied to a running gateway by `scripts/gateway_bootstrap.py` in the meteorite repository, not by editing files here.
- Every code change made for meteorite carries a `meteorite:` comment and sits in meteorite-owned files where new-api allows it, so a rebase on a new release stays small.

## Changes from upstream

None yet.
