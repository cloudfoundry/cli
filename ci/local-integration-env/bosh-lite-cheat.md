# New bosh-lite cheat sheet

If you set a different value for `$WORKSPACE` when running
`deploy_bosh_lite.sh`, change `~/workspace` to the same value for the following
commands.

## log in to bosh

```sh
export BOSH_ENVIRONMENT=vbox
export BOSH_CLIENT=admin
export BOSH_CLIENT_SECRET=$(bosh int ~/workspace/cli-lite/creds.yml --path /admin_password)
```


## ssh to bosh

```sh
bosh int ~/workspace/cli-lite/creds.yml --path /jumpbox_ssh/private_key > private_key

chmod 0600 private_key

ssh -i private_key jumpbox@192.168.50.6
```

## VirtualBox bosh-lite loses internet connectivity

Sometimes this happens after it is running for a while due to a VirtualBox bug.

```
VBoxManage natnetwork stop --netname NatNetwork
VBoxManage natnetwork start --netname NatNetwork
```

## Bringing back a dead bosh-lite

Upgrade VirtualBox to the latest version because it is more stable.

Make sure the bosh-lite VM is running in VirtualBox.

Edit `~/workspace/cli-lite/state.json` and remove the `current_manifest_sha` key.

Run `$GOPATH/src/code.cloudfoundry.org/cli/ci/local-integration-env/deploy_bosh_lite.sh`. This
will recreate the bosh-lite but not the containers.

`bosh delete-deployment -d cf --force -n`

If the deployment is locked from the resurrector:

```
bosh tasks --all
bosh cancel-task <task number from task list>
```

(Optional) To ensure that the resurrector won't try and enqueue any additional tasks:
```
bosh update-resurrection off
```

Run `$GOPATH/src/code.cloudfoundry.org/cli/ci/local-integration-env/deploy_bosh_lite.sh` again
to recreate the cf deployment. Add the `clean` argument to clean up old
bosh-lite deployment when restarting the machine puts VirtualBox in a weird
state.
