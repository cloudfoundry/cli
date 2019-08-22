---
name: Bug report
about: Create a report to help us improve

---

**Please fill out the issue checklist below and provide ALL the requested information.**

- [ ] I reviewed open and closed github issues that may be related to my problem.
- [ ] I tried updating to the latest version of the CF CLI to see if it fixed my problem.
- [ ] I attempted to run the command with `CF_TRACE=1` to help debug the issue.
- [ ] I am reporting a bug that others will be able to reproduce.
- [ ] If this is an issue for the v7 beta release, I've read through the [official docs](https://docs.cloudfoundry.org/cf-cli/v7.html) and the [release notes](https://github.com/cloudfoundry/cli/releases).

**Describe the bug and the command you saw an issue with**
Provide details on what you were trying to do (and why).

**What happened**
A clear and concise description of what happen.

**Expected behavior**
A clear and concise description of what you expected to happen.

**To Reproduce**
Steps to reproduce the behavior; include the exact CLI commands and verbose output:
1. Run `cf ...`
2. Bind a service `cf bind-service`
3. See error


**Provide more context**
- platform and shell details ( e.g. Mac OS X 10.11 iTerm)
- version of the CLI you are running
- version of the CC API Release you are on

Note: As of January 2019, we no longer support API versions older than CF Release v284/CF Deployment v1.7.0 (CAPI Release: 1.46.0 (APIs 2.100.0 and 3.35.0).

Note: In order to complete the v7 beta cf CLI in a timely matter, we develop and test against the latest CAPI release candidate. When v7 cf CLI is generally available, we will start supporting official CC API releases again.
