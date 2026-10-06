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

This first increment adds creation only. Subsequent committed TDD increments will
add list/show, bind/unbind (including job waiting and explicit restart guidance),
lifecycle operations, and explicit account-role UX.

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
