# CATS Async Broker

This directory contains an easily configured service broker that can be pushed as a CF app.

### How to push ###
-------------------
`cf push async-broker`


### Configuration API ###
------------------------
The broker includes an api that allows on-the-fly configuration.

To fetch the configuration, you can curl the following endpoint:
`curl <app_url>/config`

If you'd like to include the broker's entire state, including config, instances, and bindings:
`curl <app_url>/config/all`

To update the configuration of the broker yourself:
`curl <app_url>/config -X POST -d @data.json`

To reset the broker to its original configuration:
`curl <app_url>/config/reset -X POST`


### The Configuration File ###
------------------------------
data.json includes the basic configuration for the broker. You can update this config in many ways.

The configuration of a broker endpoint has the following form:
```json
{
  "catalog": {
    "sleep_seconds": 0,
    "status": 200,
    "body": {
      "key": "value"
    }
  }
}
```

This tells the broker to respond to a request by sleeping 0 seconds, then returning status code 200 with the JSON body
'{"key": "value"}'.

For all endpoints but the /v2/catalog endpoint, we can make the configuration dependent on the plan_id provided
in the request. This allows us to have different behavior for different situations. As an example, we might decide
that the provision endpoint (PUT /v2/service_instances/:guid) should return synchronously for one request and asynchronous
for another request. To achieve this, we configure the endpoint like this:
```json
{
  "provision": {
    "sync-plan-guid": {
      "sleep_seconds": 0,
      "status": 200,
      "body": {}
    },
    "async-plan-guid": {
      "sleep_seconds": 0,
      "status": 202,
      "body": {
        "last_operation": {
          "state": "in progress"
        }
      }
    }
  }
}
```

If we don't want to vary behavior by service plan, we can specify a default behavior:
```json
{
  "provision": {
    "default": {
      "sleep_seconds": 0,
      "status": 200,
      "body": {}
    }
  }
}
```

These behaviors are all compiled in the top-level "behaviors" key of the JSON config:

```json
{
  "behaviors": {
    "provision": { ... },
    "deprovision": { ... },
    "bind": { ... },
    "unbind": { ... },
    "fetch": { ... },
    ...
  }
}
```

##### Fetching Status #####

In the case of fetching operation status (GET /v2/service_instances/:guid/last_operation), we need to account for different behaviors
based on whether the instance operation is finished or in progress.

```json
{
  "fetch": {
    "default": {
      "in_progress": {
        "sleep_seconds": 0,
        "status": 200,
        "body": {
          "last_operation": {
            "state": "in progress"
          }
        }
      },
      "finished": {
      "sleep_seconds": 0,
        "status": 200,
        "body": {
          "last_operation": {
            "state": "succeeded"
          }
        }
      }
    }
  }
}
```

The broker will return the "in_progress" response for the endpoint until the instance's state has been requested more
times than the top-level "max_fetch_service_instance_requests" parameter. At that point, all requests to fetch in the
instance state will response with the "finished" response.

##### Asynchronous Only Behavior #####

Some brokers/ plans can only respond asynchronously to some actions. To simulate this, we can configure a plan for an action
to be 'async_only.'

```json
{
  "provision": {
    "async-plan-guid": {
      "sleep_seconds": 0,
      "async_only": true,
      "status": 200,
      "body": {}
    }
  }
}
```

If we try to provision a new service instance for the 'async-plan' without sending the 'accepts_incomplete' parameter,
the broker will respond with a 422.


### Bootstrapping ###
---------------------
The repo also includes a ruby script `setup_new_broker.rb` to bootstrap a new broker. The script does the following:
- generates a unique config for the broker in order avoid guid and name conflicts with other brokers.
- pushes the broker as an app
- registers or updates the broker
- enables service access for the broker's service

Before running the script, you must:
- Choose a CF env with `cf api`
- `cf login` with an admin user and password
- `cf target -o <some org> -s <some space>` with an org and space where you'd like to push the broker

The script also takes parameters:
- `broker_name`: Specifies the app name and route for the broker. Defaults to 'async-broker'.
- `env`: Specifies the cf environment (used only for choosing an app domain). Allowed values are 'bosh-lite',
         'tabasco', and 'a1'. Defaults to 'bosh-lite'.


### Running multiple cases ###
--------------------------
We often need to test how the CC handles many permutations of status code, body, and broker action. To make this simple, we've
written a script called run_all_cases.rb. It requires one parameter, which is a path to a CSV file. It also optionally allows
two additional parameters, a broker url and a --no-cleanup flag.

The script reads the CSV file and produces a list of test cases. For each test case, the script configures the brokers to behave
as specified, and then makes the corresponding cf cli command. The CLI output is stored in a new CSV file, which has the same
name as the input file with a "-out" suffix. (For example, acceptance.csv becomes acceptance-out.csv)

If the caller provides a broker URL, the script will use that address for its test cases, otherwise it will default to 
async-broker.bosh-lite.com.

The script also preforms setup and cleanup for each test it executes. e.g. performing an update requires a setup which
creates an instance and cleanup which deletes it.

If the user provides --no-cleanup the script will not perform a cleanup at the end of each test.


