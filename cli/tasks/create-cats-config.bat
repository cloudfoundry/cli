echo {> config.json
echo "api": "api.%DOMAIN%",>> config.json
echo "apps_domain": "%DOMAIN%",>> config.json
echo "admin_user": "%ADMIN_USER%",>> config.json
echo "admin_password": "%ADMIN_PASSWORD%",>> config.json
echo "skip_ssl_validation": true,>> config.json
echo "persistent_app_host": "persistent-app-win64",>> config.json
echo "default_timeout": 120,>> config.json
echo "cf_push_timeout": 210,>> config.json
echo "long_curl_timeout": 210,>> config.json
echo "broker_start_timeout": 330,>> config.json
echo "use_http": false,>> config.json
echo "include_v3": false>> config.json
echo }>> config.json
