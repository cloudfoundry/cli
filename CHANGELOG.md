v6.0.2

Fixed `cf push -p path/to/app.zip` on windows with zip files (eg: .zip, .war, .jar)

v6.0.1

Added purge-service-offering and migrate-service-instances commands
Added -a flag to org-users command that makes the command display all users, rather than only privileged users (#46)
Fixed a bug when manifest.yml was zero bytes
Improved error messages for commands that reference users (#79)
Fixed crash when a manifest didn't contain environment variables but there were environment variables set for the app previously
Improved error messages for commands that require an API endpoint to be set
Added timeout to all asynchronous requests
Fixed 'bad file descriptor' crash when API token expired before file upload
Added timestamps and version information to request logs when CF_TRACE is enabled
Added fallback to default log server endpoint for compatibility with older CF deployments
Improved error messages for services and target commands
Added support for URLs as arguments to create-buildpack command
Added a homebrew recipe for cf -- usage: brew install cloudfoundry-cli'
