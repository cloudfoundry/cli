package internal

var CommonHelpCategoryList = []HelpCategory{
	{
		CategoryName: "Before getting started:",
		CommandList: [][]string{
			{"config", "login", "target"},
			{"help", "logout", ""},
		},
	},

	{
		CategoryName: "Application lifecycle:",
		CommandList: [][]string{
			{"apps", "run-task", "events"},
			{"push", "logs", "set-env"},
			{"start", "ssh", "create-app-manifest"},
			{"stop", "app", ""},
			{"restart", "env", ""},
			{"restage", "scale", ""},
		},
	},

	{
		CategoryName: "Services integration:",
		CommandList: [][]string{
			{"marketplace", "create-user-provided-service"},
			{"services", "update-user-provided-service"},
			{"create-service", "create-service-key"},
			{"update-service", "delete-service-key"},
			{"delete-service", "service-keys"},
			{"service", "service-key"},
			{"bind-service", "bind-route-service"},
			{"unbind-service", "unbind-route-service"},
		},
	},

	{
		CategoryName: "Route and domain management:",
		CommandList: [][]string{
			{"routes", "delete-route", "create-domain"},
			{"domains", "map-route", ""},
			{"create-route", "unmap-route", ""},
		},
	},

	{
		CategoryName: "Space management:",
		CommandList: [][]string{
			{"spaces", "create-space", "set-space-role"},
			{"space-users", "delete-space", "unset-space-role"},
		},
	},

	{
		CategoryName: "Org management:",
		CommandList: [][]string{
			{"orgs", "create-org", "set-org-role"},
			{"org-users", "delete-org", "unset-org-role"},
		},
	},

	{
		CategoryName: "CLI plugin management:",
		CommandList: [][]string{
			{"plugins", "add-plugin-repo", "repo-plugins"},
			{"install-plugin", "list-plugin-repos", ""},
		},
	},
}
