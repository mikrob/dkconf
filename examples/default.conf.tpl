[global]
daemonize = no

[bo]
listen = 9000
listen.owner = www-data
listen.group = www-data
user = www-data
group = www-data
pm = dynamic
pm.max_children = 100
pm.start_servers = 3
pm.min_spare_servers = 1
pm.max_spare_servers = 3
pm.max_requests = 1000
pm.status_path = /fpm-status
ping.path = /fpm-ping

php_admin_value[date.timezone] = CET
