# Updates to the V7 Plugin API Interface

This document lists the differences in the objects the V7 Plugin API returns compared to V6.


## Methods that have changed

### CliCommand

With the V6 CLI plugins that used `CliCommand` to run a `cf` command were
actually calling into versions of these commands that were different to those
accessible to a user from the command line.

With the V7 CLI plugin authors will now see that command output reflects the current
implementation of those commands.

Plugins that inspect the output of `CliCommand` will need to be updated. We
recommend that if you are writing a plugin you try to depend as little as
possible on this output as it is subject to change.

#### Error handling

One significant change between older command implementations and the current
versions is that many newer commands now use `stderr` to report warnings
or failures.

We have not yet extended the plugin API to expose output on `stderr`.

When an error occurred the V6 CLI plugin API would return specific errors.
Currently the new plugin API only returns a generic error for all error cases.

### GetApp
Model changes

V6
```json
{
  "Guid": "defbb5bb-f010-4e4c-9841-60f641dd5bd4",
  "Name": "some-app",
  "BuildpackUrl": "",
  "Command": "",
  "DetectedStartCommand": "bundle exec rackup config.ru -p $PORT",
  "DiskQuota": 1024,
  "EnvironmentVars": null,
  "InstanceCount": 1,
  "Memory": 32,
  "RunningInstances": 1,
  "HealthCheckTimeout": 0,
  "State": "started",
  "SpaceGuid": "17586e2d-4df8-4879-b995-03d59a730e95",
  "PackageUpdatedAt": "2020-01-14T22:50:20Z",
  "PackageState": "STAGED",
  "StagingFailedReason": "",
  "Stack": {
    "Guid": "881b268c-d234-4c08-b5c5-50099ccb02bc",
    "Name": "cflinuxfs3",
    "Description": ""
  },
  "Instances": [
    {
      "State": "running",
      "Details": "",
      "Since": "2020-01-14T14:51:28-08:00",
      "CpuUsage": 0.0030685568679712145,
      "DiskQuota": 1073741824,
      "DiskUsage": 94228480,
      "MemQuota": 33554432,
      "MemUsage": 15935488
    }
  ],
  "Routes": [
    {
      "Guid": "93f82997-a2e8-406e-a4fd-f7c6d0d0fce5",
      "Host": "some-app",
      "Domain": {
        "Guid": "c0a63884-1260-4492-bf3c-119ed2c8a131",
        "Name": "example.com"
      },
      "Path": "",
      "Port": 0
    }
  ],
  "Services": null
}
```

V7
```json
{
  "Name": "some-app",
  "GUID": "defbb5bb-f010-4e4c-9841-60f641dd5bd4",
  "StackName": "",
  "State": "STARTED",
  "LifecycleType": "buildpack",
  "LifecycleBuildpacks": null,
  "Metadata": null,
  "ProcessSummaries": [
    {
      "GUID": "defbb5bb-f010-4e4c-9841-60f641dd5bd4",
      "Type": "web",
      "Command": "bundle exec rackup config.ru -p $PORT",
      "HealthCheckType": "port",
      "HealthCheckEndpoint": "",
      "HealthCheckInvocationTimeout": 0,
      "HealthCheckTimeout": 0,
      "Instances": 1,
      "MemoryInMB": {
        "IsSet": true,
        "Value": 32
      },
      "DiskInMB": {
        "IsSet": true,
        "Value": 1024
      },
      "Sidecars": null,
      "InstanceDetails": [
        {
          "CPU": 0.003201964487972746,
          "Details": "",
          "DiskQuota": 1073741824,
          "DiskUsage": 94228480,
          "Index": 0,
          "IsolationSegment": "",
          "MemoryQuota": 33554432,
          "MemoryUsage": 15441920,
          "State": "RUNNING",
          "Type": "web",
          "Uptime": 94000000000
        }
      ]
    },
    {
      "GUID": "bf024163-6e42-45a7-a1e1-07e14aaf899b",
      "Type": "worker",
      "Command": "bundle exec rackup config.ru",
      "HealthCheckType": "process",
      "HealthCheckEndpoint": "",
      "HealthCheckInvocationTimeout": 0,
      "HealthCheckTimeout": 0,
      "Instances": 0,
      "MemoryInMB": {
        "IsSet": true,
        "Value": 32
      },
      "DiskInMB": {
        "IsSet": true,
        "Value": 1024
      },
      "Sidecars": null,
      "InstanceDetails": null
    }
  ],
  "Routes": [
    {
      "GUID": "93f82997-a2e8-406e-a4fd-f7c6d0d0fce5",
      "SpaceGUID": "17586e2d-4df8-4879-b995-03d59a730e95",
      "DomainGUID": "c0a63884-1260-4492-bf3c-119ed2c8a131",
      "Host": "some-app",
      "Path": "",
      "DomainName": "example.com",
      "SpaceName": "some-space",
      "URL": "some-app.example.com",
      "Destinations": [
        {
          "GUID": "12c6dc1d-9a31-4deb-b0fd-5563936adae0",
          "App": {
            "GUID": "defbb5bb-f010-4e4c-9841-60f641dd5bd4",
            "Process": {
              "Type": "web"
            }
          }
        }
      ],
      "Metadata": {
        "Labels": {}
      }
    }
  ],
  "CurrentDroplet": {
    "GUID": "0110f8db-f426-4aad-87d1-4106050cd129",
    "State": "STAGED",
    "CreatedAt": "2020-01-14T22:51:13Z",
    "Stack": "cflinuxfs3",
    "Image": "",
    "Buildpacks": [
      {
        "Name": "ruby_buildpack",
        "DetectOutput": "ruby"
      }
    ]
  }
}
```

Not targeted errors:

cf6: `No org and space targeted, use 'cf target -o ORG -s SPACE' to target an org and space`

cf7:

`No organization targeted.`

`No space targeted`

### GetCurrentSpace

V6
```json
{
  "Guid": "17586e2d-4df8-4879-b995-03d59a730e95",
  "Name": "some-space"
}
```

V7
```json
{
  "Name": "some-space",
  "GUID": "17586e2d-4df8-4879-b995-03d59a730e95"
}
```

#### Error Changes

When an org isn't targeted:

V6 : Doesn't report an error, and returns an empty space object (error)

V7 : Error: no organization targeted

When a space isn't targeted:

V6 : Doesn't report an error, and returns an empty space object (error)

V7 : Error: no space targeted

### GetCurrentOrg
Model Changes

V6
```json
{
  "Guid": "d00d3542-26d2-48eb-9c39-532c10ddf487",
  "Name": "some-org",
  "QuotaDefinition": {
    "Guid": "",
    "Name": "",
    "MemoryLimit": 0,
    "InstanceMemoryLimit": 0,
    "RoutesLimit": 0,
    "ServicesLimit": 0,
    "NonBasicServicesAllowed": false
  }
}
```

V7
```json
{
  "Name": "some-org",
  "GUID": "d00d3542-26d2-48eb-9c39-532c10ddf487"
}
```
#### Error Changes

When an org isn't targeted:

V6 : Doesn't report an error, and returns an empty org object (error)

V7 : Error: no organization targeted

### GetOrg

#### Model changes
The main difference here is the removal of Quota information from GetOrg, in our
User research we found no uses of this, we also added Metadata to the V7 object

V6
```json
{
  "Guid": "d00d3542-26d2-48eb-9c39-532c10ddf487",
  "Name": "some-org",
  "QuotaDefinition": {
    "Guid": "",
    "Name": "default",
    "MemoryLimit": 102400,
    "InstanceMemoryLimit": -1,
    "RoutesLimit": 1000,
    "ServicesLimit": -1,
    "NonBasicServicesAllowed": true
  },
  "Spaces": [
    {
      "Guid": "17586e2d-4df8-4879-b995-03d59a730e95",
      "Name": "some-space"
    }
  ],
  "Domains": [
    {
      "Guid": "c0a63884-1260-4492-bf3c-119ed2c8a131",
      "Name": "example.com",
      "OwningOrganizationGuid": "",
      "Shared": true
    }
  ],
  "SpaceQuotas": null
}
```

V7

```json
{
  "Name": "some-org",
  "GUID": "d00d3542-26d2-48eb-9c39-532c10ddf487",
  "Metadata": {
    "Labels": {
      "some-key": "some-value"
    }
  },
  "Spaces": [
    {
      "Name": "some-space",
      "GUID": "17586e2d-4df8-4879-b995-03d59a730e95",
      "Metadata": {
        "Labels": null
      }
    }
  ],
  "Domains": [
    {
      "Name": "example.com",
      "GUID": "c0a63884-1260-4492-bf3c-119ed2c8a131"
    }
  ]
}
```

### GetSpace

#### Model changes
The main difference is that the V7 plugin will return only the space name and GUID, as well as its Metadata (currently existing uses of the V6 plugin are only using the `Guid` field).

V6
```json
{
  "Guid": "17586e2d-4df8-4879-b995-03d59a730e95",
  "Name": "some-space",
  "Organization": {
    "Guid": "d00d3542-26d2-48eb-9c39-532c10ddf487",
    "Name": "some-org"
  },
  "Applications": [
    {
      "Name": "some-app",
      "Guid": "defbb5bb-f010-4e4c-9841-60f641dd5bd4"
    }
  ],
  "ServiceInstances": null,
  "Domains": [
    {
      "Guid": "c0a63884-1260-4492-bf3c-119ed2c8a131",
      "Name": "example.com",
      "OwningOrganizationGuid": "",
      "Shared": true
    }
  ],
  "SecurityGroups": [
    {
      "Name": "public_networks",
      "Guid": "8a1f1ebf-41a4-4739-ac9f-e9a52e3e583c",
      "Rules": [
        {
          "destination": "0.0.0.0-9.255.255.255",
          "protocol": "all"
        }
      ]
    }
  ],
  "SpaceQuota": {
    "Guid": "",
    "Name": "",
    "MemoryLimit": 0,
    "InstanceMemoryLimit": 0,
    "RoutesLimit": 0,
    "ServicesLimit": 0,
    "NonBasicServicesAllowed": false
  }
}
```

V7

```json
{
  "Name": "some-space",
  "GUID": "17586e2d-4df8-4879-b995-03d59a730e95",
  "Metadata": {
    "Labels": {
      "some-key": "some-value"
    }
  }
}
```

### GetApps
#### Model Changes

V6
```json
[
  {
    "Name": "some-app",
    "Guid": "defbb5bb-f010-4e4c-9841-60f641dd5bd4",
    "State": "started",
    "TotalInstances": 1,
    "RunningInstances": 1,
    "Memory": 32,
    "DiskQuota": 1024,
    "Routes": [
      {
        "Guid": "93f82997-a2e8-406e-a4fd-f7c6d0d0fce5",
        "Host": "some-app",
        "Domain": {
          "Guid": "c0a63884-1260-4492-bf3c-119ed2c8a131",
          "Name": "example.com",
          "OwningOrganizationGuid": "",
          "Shared": false
        }
      }
    ]
  }
]
```

V7
```json
[
  {
    "Name": "some-app",
    "GUID": "defbb5bb-f010-4e4c-9841-60f641dd5bd4",
    "StackName": "cflinuxfs3",
    "State": "STARTED",
    "LifecycleType": "buildpack",
    "LifecycleBuildpacks": null,
    "Metadata": {
      "Labels": {
        "some-key": "some-value"
      }
    }
  }
]
```


#### Not-targeted errors

When no space is targeted:

V6: Returns error: `No org and space targeted, use 'cf target -o ORG -s SPACE' to target an org and space`

V7: Returns error: `no space targeted`

### Username

No interface changes

#### Not logged in behaviour

V6: Returns empty string with no error

V7: Returns empty string with error `not logged in`

### GetSpaces
#### Model Changes

V6:
```json
[
  {
    "Guid": "17586e2d-4df8-4879-b995-03d59a730e95",
    "Name": "some-space"
  }
]
```

V7:
```json
[
  {
    "Name": "some-space",
    "GUID": "17586e2d-4df8-4879-b995-03d59a730e95",
    "Metadata": {
      "Labels": {
        "some-key": "some-value"
      }
    }
  }
]
```

### IsSSLDisabled (Renamed to IsSkipSSLValidation in V7)

The only difference in this method is that it was renamed from the V6
`IsSSLDisabled` to the V7 `IsSkipSSLValidation`


## Methods that have not changed

AccessToken
ApiEndpoint
IsLoggedIn

