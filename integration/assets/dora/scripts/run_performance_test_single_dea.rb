#!/usr/bin/env ruby
DEFAULT_APPLICATION_URL = "dora.sunset.cf-app.com"
app_url = ENV['EXPERIMENT_APP_URL'] || DEFAULT_APPLICATION_URL

dea_instances = Hash.new { |hash, key| hash[key] = [] }
while gets()
  instance_info = $_.split
  dea_instances[instance_info[2]].push({index: instance_info[1], cookie_jar: instance_info[0]})
end

populous_dea = dea_instances.keys.sort_by { |key| dea_instances[key].length }.last
raise "No DEA with two instances" unless dea_instances[populous_dea].length >= 2

app_instances = dea_instances[populous_dea][0..1]

STDERR.puts "Running experiment against #{app_url} on dea #{populous_dea}"
STDERR.puts "log/sleep output from app instance #{app_instances[0][:index]}"
STDERR.puts "loglines output from app instance #{app_instances[1][:index]}"

cjar_0 = app_instances[0][:cookie_jar]
cjar_1 = app_instances[1][:cookie_jar]

pid = Process.spawn("curl #{app_url}/log/sleep/100000 -b #{cjar_0}")
sleep 5

system("curl #{app_url}/loglines/1000/TAG1TAG -b #{cjar_1}")
sleep 5

system("curl #{app_url}/log/stop -b #{cjar_0}")

Process.wait pid
STDERR.puts "all done"