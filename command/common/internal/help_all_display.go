package internal

var HelpCategoryList = []HelpCategory{
	{
		CategoryName: "GETTING STARTED:",
		CommandList: [][]string{
			{"help", "version", "login", "logout", "passwd", "target"},
			{"api", "auth"},
		},
	},
	{
		CategoryName: "APPS:",
		CommandList: [][]string{
			{"apps", "app"},
			{"push", "scale", "delete", "rename"},
			{"start", "stop", "restart", "restage", "restart-app-instance"},
			{"run-task", "tasks", "terminate-task"},
			{"events", "files", "logs"},
			{"env", "set-env", "unset-env"},
			{"stacks", "stack"},
			{"copy-source", "create-app-manifest"},
			{"get-health-check", "set-health-check", "enable-ssh", "disable-ssh", "ssh-enabled", "ssh"},
		},
	},
	{
		CategoryName: "SERVICES:",
		CommandList: [][]string{
			{"marketplace", "services", "service"},
			{"create-service", "update-service", "delete-service", "rename-service"},
			{"create-service-key", "service-keys", "service-key", "delete-service-key"},
			{"bind-service", "unbind-service"},
			{"bind-route-service", "unbind-route-service"},
			{"create-user-provided-service", "update-user-provided-service"},
		},
	},
	{
		CategoryName: "ORGS:",
		CommandList: [][]string{
			{"orgs", "org"},
			{"create-org", "delete-org", "rename-org"},
		},
	},
	{
		CategoryName: "SPACES:",
		CommandList: [][]string{
			{"spaces", "space"},
			{"create-space", "delete-space", "rename-space"},
			{"allow-space-ssh", "disallow-space-ssh", "space-ssh-allowed"},
		},
	},
	{
		CategoryName: "DOMAINS:",
		CommandList: [][]string{
			{"domains", "create-domain", "delete-domain", "create-shared-domain", "delete-shared-domain"},
			{"router-groups"},
		},
	},
	{
		CategoryName: "ROUTES:",
		CommandList: [][]string{
			{"routes", "create-route", "check-route", "map-route", "unmap-route", "delete-route", "delete-orphaned-routes"},
		},
	},
	{
		CategoryName: "NETWORK POLICIES:",
		CommandList: [][]string{
			{"network-policies", "add-network-policy", "remove-network-policy"},
		},
	},
	{
		CategoryName: "BUILDPACKS:",
		CommandList: [][]string{
			{"buildpacks", "create-buildpack", "update-buildpack", "rename-buildpack", "delete-buildpack"},
		},
	},
	{
		CategoryName: "USER ADMIN:",
		CommandList: [][]string{
			{"create-user", "delete-user"},
			{"org-users", "set-org-role", "unset-org-role"},
			{"space-users", "set-space-role", "unset-space-role"},
		},
	},
	{
		CategoryName: "ORG ADMIN:",
		CommandList: [][]string{
			{"quotas", "quota", "set-quota"},
			{"create-quota", "delete-quota", "update-quota"},
			{"share-private-domain", "unshare-private-domain"},
		},
	},
	{
		CategoryName: "SPACE ADMIN:",
		CommandList: [][]string{
			{"space-quotas", "space-quota"},
			{"create-space-quota", "update-space-quota", "delete-space-quota"},
			{"set-space-quota", "unset-space-quota"},
		},
	},
	{
		CategoryName: "SERVICE ADMIN:",
		CommandList: [][]string{
			{"service-auth-tokens", "create-service-auth-token", "update-service-auth-token", "delete-service-auth-token"},
			{"service-brokers", "create-service-broker", "update-service-broker", "delete-service-broker", "rename-service-broker"},
			{"migrate-service-instances", "purge-service-offering", "purge-service-instance"},
			{"service-access", "enable-service-access", "disable-service-access"},
		},
	},
	{
		CategoryName: "SECURITY GROUP:",
		CommandList: [][]string{
			{"security-group", "security-groups", "create-security-group", "update-security-group", "delete-security-group", "bind-security-group", "unbind-security-group"},
			{"bind-staging-security-group", "staging-security-groups", "unbind-staging-security-group"},
			{"bind-running-security-group", "running-security-groups", "unbind-running-security-group"},
		},
	},
	{
		CategoryName: "ENVIRONMENT VARIABLE GROUPS:",
		CommandList: [][]string{
			{"running-environment-variable-group", "staging-environment-variable-group", "set-staging-environment-variable-group", "set-running-environment-variable-group"},
		},
	},
	{
		CategoryName: "ISOLATION SEGMENTS:",
		CommandList: [][]string{
			{"isolation-segments", "create-isolation-segment", "delete-isolation-segment", "enable-org-isolation", "disable-org-isolation", "set-org-default-isolation-segment", "reset-org-default-isolation-segment", "set-space-isolation-segment", "reset-space-isolation-segment"},
		},
	},
	{
		CategoryName: "FEATURE FLAGS:",
		CommandList: [][]string{
			{"feature-flags", "feature-flag", "enable-feature-flag", "disable-feature-flag"},
		},
	},
	{
		CategoryName: "ADVANCED:",
		CommandList: [][]string{
			{"curl", "config", "oauth-token", "ssh-code"},
		},
	},
	{
		CategoryName: "ADD/REMOVE PLUGIN REPOSITORY:",
		CommandList: [][]string{
			{"add-plugin-repo", "remove-plugin-repo", "list-plugin-repos", "repo-plugins"},
		},
	},
	{
		CategoryName: "ADD/REMOVE PLUGIN:",
		CommandList: [][]string{
			{"plugins", "install-plugin", "uninstall-plugin"},
		},
	},
}

var ExperimentalHelpCategoryList = []HelpCategory{
	{
		CategoryName: "APPS (experimental):",
		CommandList: [][]string{
			{"v3-apps", "v3-app", "v3-create-app"},
			{"v3-push", "v3-scale", "v3-delete"},
			{"v3-start", "v3-stop", "v3-restart", "v3-stage", "v3-restart-app-instance"},
			{"v3-droplets", "v3-set-droplet"},
			{"v3-set-env", "v3-unset-env"},
			{"v3-get-health-check", "v3-set-health-check"},
			{"v3-packages", "v3-create-package"},
		},
	},
}
