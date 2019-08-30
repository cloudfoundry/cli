# CLI Concourse

## Setup MacOS Concourse worker

- Copy `./com.pivotal.ConcourseWorker.plist` to `/Library/LaunchDaemons`
- Copy `./bin/concourse_worker` to `/Users/pivotal/bin/concourse_worker`
- Load service `sudo launchctl load -w /Library/LaunchDaemons/com.pivotal.StartupConcourse.plist`
- Start service `launchctl start com.pivotal.StartupConcourse`
- Check system logs using `lnav`
- Check worker logs `sudo tail -f /usr/local/var/log/concourse-worker.std*.log`
