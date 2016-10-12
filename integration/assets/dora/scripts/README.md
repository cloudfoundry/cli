Scripts to assist in running loggregator performance charaterization experiments

1. Start a drain server somewhere

    One option is to use `socat` on one of the router job vms. 
    ```
bosh ssh router_z1
sudo -i
apt-get install socat
mkdir -p /var/vcap/sys/logs/drain
socat -u TCP4-LISTEN:4567,reuseaddr,fork - | \
 ruby -ne 'print Time.now.strftime("%FT%T.%N%:z") + " " + $_; $stdout.flush' | \
 tee /var/vcap/sys/logs/drain/messages.log
```

1. Push dora, and bind the syslog drain url to it

    ```
cf push dora
cf delete-service sc -f
cf cups sc -l syslog://10.10.16.15:4567
cf bind-service dora sc
cf restart dora
```

1. Use the scale_dora script to achieve n instances evenly distributed on the DEAs

    ```
scale_dora -i 8
```

1. Use the scripts to create cookie_jar files for sticky sessions to each running instance, and map those files to ip addresses and dora indicies

    ```
mkdir cookie_jars
get_instance_cookie_jars.sh -e 8 -m 40 -d cookie_jars
map_cookie_jars_to_instances.rb > map.out
```

1. Run the experiment script

    ```
run_performance_test_single_dea.rb < map.out
```

1. Grab the messages.log file from the drain server.

    Since the drain server is recording the messages to a file that ends in `.log`, and is somewhere under `/var/vcap/sys/logs`, we can use `bosh logs` to download that file.

    ```
bosh logs router_z1
tar xvfz router_z1*.tgz
```

1. Perform desired analysis on the messages.log file

    Maybe look for missing lines, propagation delays, etc. 
