#!/usr/bin/env ruby
DEFAULT_APPLICATION_URL = "dora.sunset.cf-app.com"
app_url = ENV['EXPERIMENT_APP_URL'] || DEFAULT_APPLICATION_URL

dea_instances = Hash.new { |hash, key| hash[key] = [] }
while gets()
  instance_info = $_.split
  dea_instances[instance_info[2]].push({index: instance_info[1], cookie_jar: instance_info[0]})
end

populous_dea = dea_instances.keys.sort_by { |key| dea_instances[key].length }.last
raise "No DEA with 8 instances" unless dea_instances[populous_dea].length >= 8

app_instances = dea_instances[populous_dea][0..7]

STDERR.puts "Running experiment against #{app_url} on dea #{populous_dea}"
app_instances[0..6].each do |i|
	STDERR.puts "log/sleep output from app instance #{i[:index]}"
end
STDERR.puts "loglines output from app instance #{app_instances[7][:index]}"

cjars = app_instances.map { |i| i[:cookie_jar]}

pid=[]

8.times do |background_count|
	STDERR.puts "starting experiment with #{background_count} log/sleep"
        STDERR.puts "placing marker line"
        system("curl -s #{app_url}/loglines/1/MARKER -b #{cjars[7]}")
        sleep 1

	background_count.times do |b|
		pid[b] = Process.spawn("curl -s #{app_url}/log/sleep/1000 -b #{cjars[b]}")
	end
	sleep 5

	100.times do |iteration_count|
		system("curl -s #{app_url}/loglines/1000/BG#{background_count}ITER#{iteration_count} -b #{cjars[7]}")
		sleep 2
	end
	sleep 5


	background_count.times do |b|
		pid[b] = Process.spawn("curl -s #{app_url}/log/stop -b #{cjars[b]}")
	end

	background_count.times do |b|
		Process.wait pid[b]
	end
	sleep 5

	total_line_count = 0

	background_count.times do |b|
		line_count=`curl -s #{app_url}/log/sleep/count -b #{cjars[b]}`
		puts "#{line_count} lines for background job #{b}"
		total_line_count = total_line_count + line_count.to_i
	end	

	puts "total line count is #{total_line_count}"
end

STDERR.puts "all done"
