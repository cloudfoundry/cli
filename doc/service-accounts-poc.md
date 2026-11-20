# Service-account commands (RFC lab)

This branch implements the space-owned service-account proposal against the lab's
CAPI extensions. The API is not yet released upstream; unsupported foundations
return a CAPI error. A minimum API version will be assigned when the API is
standardized rather than treating an existing version as feature support.

## Create an account

```sh
cf target -o poc -s service-accounts-e2e
cf create-service-account shared-worker --description "Background workers"
```

Creation requires login, a targeted organization and space, and CAPI account
management permission (space manager or platform admin in the local profile).
The request uses the targeted space as the immutable owner. Names are immutable,
foundation-unique, and permanently reserved after deletion; CAPI validates them.

Output shows the platform-returned client ID, certificate DNS SAN, and provisioning
status. Creating the account does not provision its UAA client, bind an app, or
grant resource roles. Provisioning occurs on first authorized app assignment.

## List accounts

Run `cf service-accounts` to list accounts owned by the targeted space. The table
shows name, enabled state, provisioning status, client ID and certificate DNS SAN.
All result pages are retrieved; API warnings and errors are preserved.

## Lifecycle and app assignment

```sh
cf service-account shared-worker
cf bind-service-account my-app shared-worker
cf unbind-service-account my-app
cf disable-service-account shared-worker
cf enable-service-account shared-worker
cf delete-service-account shared-worker
```

All name resolution is scoped to the targeted space. Bind/unbind and account
lifecycle commands wait for asynchronous CAPI jobs before printing success.
Bind/unbind update the desired assignment only: restart the app explicitly to
apply the new launch identity. Existing certificates and tokens are not revoked.
Binding does not grant resource roles.

Disabling prevents new token issuance once reconciliation completes; existing
tokens and certificate-only access persist until their normal expiry. Deletion
requires the account to be unused and retains its name reservation permanently.
Deletion asks for confirmation unless `-f` is supplied; this flag skips the prompt,
not CAPI's in-use checks. Explicit account-role UX remains a follow-up.

Creation, listing and all lifecycle commands are listed in both `cf help` and
`cf help -a`. For this checkout, use the built custom binary or `go run ./main.go`
from `cli/`; the installed stock CLI does not contain these local extensions.

## Development

From the lab workspace root:

```sh
devbox run -- env -C "$PWD/cli" go test ./api/cloudcontroller/ccv3 ./actor/v7action ./command/v7 ./command/common -ginkgo.focus='Service Account|create-service-account'
devbox run -- env -C "$PWD/cli" go test ./resources ./api/cloudcontroller/ccv3 ./actor/v7action ./command/v7 ./command/common
devbox run -- env -C "$PWD/cli" go build -o ../.local/cf-service-accounts .
./.local/cf-service-accounts help create-service-account
```

Use separate committed RED/GREEN cycles. Generate changed interface fakes using
the repository-pinned Counterfeiter; do not hand-edit generated output.
