# Updates to the v7 Plugin API Interface

This document lists the differences in the objects the V7 Plugin API returns compared to V6.

## Methods that have changed

### GetExample

GetExample returned an "example" string on success and an error string on error in V6

In V7 it now returns "cool example" with no change to the error text

### GetApp
Model changes (TODO)

V6
```
{
        "Guid": STRING,
        "Name": STRING,
        "BuildpackUrl": STRING,
        "Command": STRING,
        "DetectedStartCommand": STRING,
        "DiskQuota": 1024,
        "EnvironmentVars": HASH,
        "InstanceCount": 1,
        "Memory": 32,
        "RunningInstances": 1,
        "HealthCheckTimeout": 0,
        "State": "started",
        "SpaceGuid": STRING,
        "PackageUpdatedAt": STRING,
        "PackageState": "STAGED",
        "StagingFailedReason": STRING,
        "Stack": {
                "Guid": STRING,
                "Name": STRING,
                "Description": STRING
        },
        "Instances": [
                STATS_HASH
        ],
        "Routes": [
                {
                        "Guid": STRING,
                        "Host": "dora",
                        "Domain": {
                                "Guid": STRING,
                                "Name": "ancient-twister.lite.cli.fun"
                        },
                        "Path": STRING,
                        "Port": 0
                }
        ],
        "Services": null
}
```

V7
```
{
        "Name": STRING,
        "GUID": STRING,
        "StackName": "",
        "State": "STARTED",
        "LifecycleType": "buildpack",
        "LifecycleBuildpacks": null,
        "Metadata": null,
        "ProcessSummaries": [
                {
                        "GUID": STRING,
                        "Type": "web",
                        "Command": "bundle exec rackup config.ru -p $PORT",
                        "HealthCheckType": STRING,
                        "HealthCheckEndpoint": STRING,
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
                                        "CPU": 0.002718184649961429,
                                        "Details": "",
                                        "DiskQuota": 1073741824,
                                        "DiskUsage": 115396608,
                                        "Index": 0,
                                        "IsolationSegment": "",
                                        "MemoryQuota": 33554432,
                                        "MemoryUsage": 14675968,
                                        "State": "RUNNING",
                                        "Type": "web",
                                        "Uptime": 2437000000000
                                }
                        ]
                }
        ],
        "Routes": [
                {
                        "GUID": STRING,
                        "SpaceGUID": STRING,
                        "DomainGUID": STRING,
                        "Host": "dora",
                        "Path": STRING,
                        "DomainName": STRING,
                        "SpaceName": STRING,
                        "URL": STRING
                }
        ],
        "CurrentDroplet": {
                "GUID": STRING,
                "State": "STAGED",
                "CreatedAt": STRING,
                "Stack": STRING,
                "Image": STRING,
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

cf7: `Application 'APP' not found.`

### GetCurrentSpace

v6
```
{SpaceFields:
  {Guid:a53cf02c-4d18-4955-b321-b8c45e33df1e 
   Name:space
   }
}
```

v7
```
{Name:space GUID:a53cf02c-4d18-4955-b321-b8c45e33df1e}
```

#### Error Changes

When an org isn't targeted:

v6 : Doesn't report an error, and returns an empty space object (error)

v7 : Error: no organization targeted

When a space isn't targeted:

v6 : Doesn't report an error, and returns an empty space object (error)

v7 : Error: no space targeted

### GetCurrentOrg
Model Changes

v6
```
{
    OrganizationFields:{
        Guid:STRING 
        Name:STRING 
        QuotaDefinition: {
            Guid: STRING: 
            MemoryLimit:0 
            InstanceMemoryLimit:0 
            RoutesLimit:0 
            ServicesLimit:0 
            NonBasicServicesAllowed:false}
    }
}
```

v7
```
{
    Name:org 
    GUID:2e2eb386-34a9-4b29-ad49-365a57bdca6c
}
```
Error Changes?

When an org isn't targeted:

v6 : Doesn't report an error, and returns an empty org object (error)

v7 : Error: no organization targeted

### GetApps
#### Model Changes

V6
```
[
    {
        Name:pora Guid:d8e2ed07-804b-4ed1-b2e6-21953d13c0f7 
        State:started 
        TotalInstances:1 
        RunningInstances:1 
        Memory:256 
        DiskQuota:1024 
        Routes:[
            {
                Guid:7248a64a-e98e-436c-851a-4fb9603b3e84 
                Host:pora 
                Domain:{
                    Guid:177b07ba-b9e5-4189-b801-8a59cdf79021 
                    Name:quasar-scarer.capi.land 
                    OwningOrganizationGuid: 
                    Shared:false
                }
            }
        ]
    },
    ...
]
```

V7
```
[
    {
        Name:pora 
        GUID:d8e2ed07-804b-4ed1-b2e6-21953d13c0f7 
        StackName:cflinuxfs3 
        State:STARTED 
        LifecycleType:buildpack 
        LifecycleBuildpacks:[go_buildpack] 
        Metadata: {Labels:map[]}
    },
    ...
]
```


#### Not-targeted errors

When no space is targeted:

v6: Returns error: `No org and space targeted, use 'cf target -o ORG -s SPACE' to target an org and space`

v7: Returns error: `no space targeted`

### Username

No interface changes

#### Not logged in behaviour

v6: Returns empty string with no error

v7: Returns empty string with error `not logged in`

## Methods that have not changed

AccessToken
ApiEndpoint
IsLoggedIn


## Other changes
When the CLI is not targeted at an API in v6, plugin methods will just return null objects with an error attatched, in v7 the plugin command will error in the CLI code without executing any of the plugin specific code. This will display the error in stderr like other CLI commands do.


Not logged in errors? (TODO)
