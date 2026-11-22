# meteorite gateway

This is the meteorite fork of [QuantumNous/new-api](https://github.com/QuantumNous/new-api), the gateway of meteorite.run (stage 1).

- Based on release `v1.0.0-rc.51` (git remote `upstream`). Record the new base here whenever the fork is rebased.
- License: AGPL-3.0, the same as new-api. The full text is in `LICENSE`; the running web UI links to this repository from its footer, so users can get the source.
- Settings that do not need code (name, logo, footer, sign-up options) are applied to a running gateway by `scripts/gateway_bootstrap.py` in the meteorite repository, not by editing files here.
- Every code change made for meteorite carries a `meteorite:` comment and sits in meteorite-owned files where new-api allows it, so a rebase on a new release stays small.

## Changes from upstream

- `service/meteorite_billing.go`, `setting/meteorite_price.go`: input tokens are billed as reported minus the channel's average injection (never below 10, never above the reported count; D-063), and the price is a percentage of the official price per model, 50% by default (R-013). Option `MeteoritePricePercent`, for example `{"claude-sonnet-5-5": 60}`.
- `relaykit/dto/channel_settings.go`: field `meteorite_injection_tokens`, written by the verification backend.
- `service/text_quota.go`, `relay/helper/price.go`, `model/option.go`: one `meteorite:` hook each.
- `controller/relay.go`: a request refused for lack of balance answers 402.
- `model/meteorite_cooldown.go`: a channel that failed a user request rests for 60 seconds, in both channel selection paths (R-015); hooked from `controller/relay.go`. Retries and the ban switch are settings (`RetryTimes` 2, automatic disable off), applied by `scripts/gateway_bootstrap.py`.
- `model/meteorite_seal.go`: API keys and channel keys are stored sealed (AES-256-GCM, key `METEORITE_SEAL_KEY`), sealed again after a restart if any plain one is found (R-023, R-005). Hooks on `Token` and `Channel`, and one lookup in `model/token.go`.
- `service/meteorite_login_limit.go`, `controller/user.go`: five failed sign-ins for an account from one address block it for 15 minutes (R-023).
- `web/src/features/wallet/components/meteorite-topup-link.tsx`: the link to the test top-up page, shown only while the meteorite service says the test top-up is on (D-069).
