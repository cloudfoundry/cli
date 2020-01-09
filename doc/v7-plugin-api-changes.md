# Updates to the V7 Plugin API Interface

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

cf7:

`No organization targeted.`

`No space targeted`

### GetCurrentSpace

V6
```
{SpaceFields:
  {Guid:a53cf02c-4d18-4955-b321-b8c45e33df1e 
   Name:space
   }
}
```

V7
```
{Name:space GUID:a53cf02c-4d18-4955-b321-b8c45e33df1e}
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

V7
```
{
    Name:org 
    GUID:2e2eb386-34a9-4b29-ad49-365a57bdca6c
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
```
  Guid:1a873322-a1cb-47aa-ad81-5b7460941910 Name:system 
  QuotaDefinition:{
    Guid:
    Name:default
    MemoryLimit:102400 
    InstanceMemoryLimit:-1
    RoutesLimit:1000
    ServicesLimit:-1
    NonBasicServicesAllowed:true
  }
  Spaces:[{
    Guid:00954f6d-fd0f-47f7-80e8-a597a47df9df Name:test-org
    }
  ]
  Domains:[{
    Guid:e672e331-def1-418b-ad4c-428177de353d
    Name:frost-dagger.lite.cli.fun OwningOrganizationGuid: Shared:true
  }]
  SpaceQuotas:[{SpaceQuoteObject}]}
```
V7

```
{
   Org: {
     Name: system GUID: 1a873322-a1cb-47aa-ad81-5b7460941910
   }
   Metadata: {
     Labels: map[
        label: { Value: test IsSet: true }
        fun: { Value: true IsSet: true }
      ]
   }
   Spaces: [
     {
       Name: test-org GUID: 00954f6d-fd0f-47f7-80e8-a597a47df9df
     }
   ]
   Domains: [
     {
       Name: frost-dagger.lite.cli.fun
       GUID: e672e331-def1-418b-ad4c-428177de353d
     }
   ]
}
```

### GetSpace

#### Model changes
The main difference here is the removal of Quota information from GetOrg, in our
User research we found no uses of this, we also added Metadata to the V7 object

V6
```
{
  Guid:a53cf02c-4d18-4955-b321-b8c45e33df1e Name:myspace 
  Organization:{Guid:2e2eb386-34a9-4b29-ad49-365a57bdca6c Name:tom}
  Applications:[{Name:pora Guid:d8e2ed07-804b-4ed1-b2e6-21953d13c0f7} {Name:zora Guid:85706f87-14b9-4e9b-b9eb-7f3aee2fe9a8}]
  ServiceInstances:[]
  Domains:[{
    Guid:e672e331-def1-418b-ad4c-428177de353d
    Name:frost-dagger.lite.cli.fun OwningOrganizationGuid: Shared:true
  }]
  SecurityGroups:[
    { 
    Name:public_networks 
    Guid:3b4b1860-446c-4827-ac46-41c69a53868c 
    Rules:[
      map[destination:0.0.0.0-9.255.255.255 protocol:all] 
      map[destination:11.0.0.0-169.253.255.255 protocol:all] 
      map[destination:169.255.0.0-172.15.255.255 protocol:all] 
      map[destination:172.32.0.0-192.167.255.255 protocol:all] 
      map[destination:192.169.0.0-255.255.255.255 protocol:all]]
    } 
    {
       Name:dns 
       Guid:b10e310c-e3c1-4882-9835-58b413b9c9b8 
       Rules:[map[destination:0.0.0.0/0 ports:53 protocol:tcp] map[destination:0.0.0.0/0 ports:53 protocol:udp]]
       }
   ]
 SpaceQuota: {
   Guid: 
   Name: 
   MemoryLimit:0 
   InstanceMemoryLimit:0 
   RoutesLimit:0 ServicesLimit:0 
   NonBasicServicesAllowed:false
   }
  }
}
```
V7

```
{
   Space: {
     Name: myspace GUID: 1a873322-a1cb-47aa-ad81-5b7460941910
   }
   Org: {
     Name: toim GUID: 3a873322-a1cb-47aa-ad81-5b7460941910
   }
   Metadata: {
     Labels: map[
        label: { Value: test IsSet: true }
        fun: { Value: true IsSet: true }
      ]
   }
   Apps: [
     {
       Name: shmora GUID: 00954f6d-fd0f-47f7-80e8-a597a47df9df
     }
   ]
   Domains: [
     {
       Name: frost-dagger.lite.cli.fun
       GUID: e672e331-def1-418b-ad4c-428177de353d
     }
   ]
}

```

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
  Name: myspace
  Guid: myguid
 },...
]
```

V7:
```
[
 {
  Name: myspace
  GUID: myguid
  Metadata: {
     Labels: map[key:{Value: avalue}]
  }
 },...
]
```


### IsSSLDisabled (Renamed to IsSkipSSLValidation in V7)

The only difference in this method is that it was renamed from the V6
`IsSSLDisabled` to the V7 `IsSkipSSLValidation`

## Methods that have not changed

AccessToken
ApiEndpoint
IsLoggedIn


## Other changes
When the CLI is not targeted at an API in V6, plugin methods will just return null objects with an error attatched, in V7 the plugin command will error in the CLI code without executing any of the plugin specific code. This will display the error in stderr like other CLI commands do.


Not logged in errors? (TODO)
