#!/usr/bin/env ruby
DEFAULT_APPLICATION_URL = "dora.sunset.cf-app.com"
app_url = ENV['EXPERIMENT_APP_URL'] || DEFAULT_APPLICATION_URL
fg_count = "1000"

dea_instances = Hash.new { |hash, key| hash[key] = [] }
while gets()
  instance_info = $_.split
  dea_instances[instance_info[2]].push({index: instance_info[1], cookie_jar: instance_info[0]})
end

# get 2 deas with 8 instances each
sorted_deas = dea_instances.keys.sort_by { |key| dea_instances[key].length }
populous_dea0 = sorted_deas[-1]
populous_dea1 = sorted_deas[-2]
raise "Need 2 DEAs with 8 instances" unless (dea_instances[populous_dea0].length >= 8 && dea_instances[populous_dea1].length >= 8)

stable_instances = dea_instances[populous_dea0][0..7]

app_instances = dea_instances[populous_dea1][0..7]

# display info on which instances have which role in the experiment
STDERR.puts "Running experiment against #{app_url} on dea #{populous_dea0} and dea #{populous_dea1}"
(stable_instances[0..6] + app_instances).each do |i|
	STDERR.puts "log/sleep output from app instance #{i[:index]}"
end
STDERR.puts "loglines output from app instance #{stable_instances[7][:index]}"


app_cjars = app_instances.map { |i| i[:cookie_jar]}
stable_cjars = stable_instances.map { |i| i[:cookie_jar]}
fg_cjar = stable_cjars[7]


# run 1 fg, 7 bg on first dea, and run 1-8 bg on second dea
9.times do |background_count|
  pid=[]
	STDERR.puts "starting experiment with #{background_count} log/sleep on second dea"
  STDERR.puts "placing marker line"
  system("curl -s #{app_url}/loglines/1/MARKER -b #{fg_cjar}")
        sleep 1

	background_count.times do |b|
		pid << Process.spawn("curl -s #{app_url}/log/sleep/1000 -b #{app_cjars[b]}")
  end
  7.times do |b|
    pid << Process.spawn("curl -s #{app_url}/log/sleep/1000 -b #{stable_cjars[b]}")
  end
	sleep 5

	100.times do |iteration_count|
		system("curl -s #{app_url}/loglines/1000/BG#{background_count}ITER#{iteration_count} -b #{fg_cjar}")
		sleep 2
	end
	sleep 5

  background_count.times do |b|
    pid << Process.spawn("curl -s #{app_url}/log/stop -b #{app_cjars[b]}")
  end
  7.times do |b|
    pid << Process.spawn("curl -s #{app_url}/log/stop -b #{stable_cjars[b]}")
  end

	pid.each do |p|
		Process.wait p
	end
	sleep 5

	total_line_count = 0

	background_count.times do |b|
		line_count=`curl -s #{app_url}/log/sleep/count -b #{app_cjars[b]}`
		puts "#{line_count} lines for background job #{b}"
		total_line_count = total_line_count + line_count.to_i
  end
  puts "total line count for background jobs is #{total_line_count}"

  total_line_count = 0
  7.times do |b|
    line_count=`curl -s #{app_url}/log/sleep/count -b #{stable_cjars[b]}`
    puts "#{line_count} lines for stable job #{b}"
    total_line_count = total_line_count + line_count.to_i
  end

	puts "total line count for stable jobs is #{total_line_count}"
end

STDERR.puts "all done"
