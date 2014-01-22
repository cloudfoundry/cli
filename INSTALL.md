## Windows

1. Download the CLI from github: https://github.com/cloudfoundry/cli/releases
2. Extract the zip file.
3. Create a folder in C:\Program Files\, named "Cloud Foundry"
4. Move `cf` to C:\Program Files\Cloud Foundry\
5. Set your %PATH% to include C:\Program Files\Cloud Foundry [(see instructions)](http://www.wikihow.com/Create-a-Custom-Windows-Command-Prompt)
  1. Right-click My Computer > Properties
  2. Click on Advanced system settings
  3. Click on Environment Variables
  4. Click on "Path" in the System Variables list
  5. Click Edit
  6. Append C:\Program Files\Cloud Foundry\ to the Variable value separated by a semicolon
  7. Click OK
  8. Click OK
6. Open up the command prompt and type `cf`
7. You should see the CLI help if everything is successful

## Mac OSX and Linux

1. Download the CLI from github: https://github.com/cloudfoundry/cli/releases
2. Extract the tgz file.
3. Move `cf` to /usr/local/bin
4. Confirm /usr/local/bin is in your PATH by typing `echo $PATH` at the command line
5. Type `cf` at the command line
6. You should see the CLI help if everything is successful
