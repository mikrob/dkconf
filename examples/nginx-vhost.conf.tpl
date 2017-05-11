server {
    listen      80;
    root        /app/web;

    server_name {{ range $idx, $elem := .SecondaryServersNames}}{{ $elem }} {{ end }};
    access_log  /var/log/nginx/access.log main;
    error_log   /var/log/nginx/error.log;

    {{ .TrucBidule }}

    {{ .Fqdn }}

    location / {
        if ($http_x_forwarded_proto != "https") {
          return 301 https://www.{{ .Fqdn }}$request_uri;
        }
        if ($host != "www.{{ .MainServerName }}") {
          return 301 https://www.{{ .Fqdn }}$request_uri;
        }
        try_files $uri /app.php$is_args$args;
    }

    location ~ ^/app\.php(/|$) {
        fastcgi_pass            fpm:9000;
        fastcgi_split_path_info ^(.+\.php)(/.*)$;
        include                 fastcgi_params;
        fastcgi_param           SCRIPT_FILENAME $document_root$fastcgi_script_name;
        fastcgi_param           HTTPS off;
        fastcgi_param  SYMFONY__APP_DEFAULT_TENANT  premium;

        internal;
        gzip                    on;
        gzip_comp_level         3;

        {{ if .CorsEnabled }}
        if ($request_method = 'OPTIONS') {
          add_header 'Access-Control-Allow-Origin' '*' always;
          add_header 'Access-Control-Allow-Credentials' 'true' always;
          add_header 'Access-Control-Allow-Methods' 'GET, POST, OPTIONS, DELETE, HEAD, PUT, PATCH' always;
          add_header 'Access-Control-Expose-Headers' 'DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Api-User,X-Api-Client,X-UA-Device-Category,X-UA-Device' always;
          add_header 'Access-Control-Allow-Headers' 'DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Api-User,X-Api-Client,X-UA-Device-Category,X-UA-Device' always;
          add_header 'Access-Control-Max-Age' 1728000 always;
          add_header 'Content-Type' 'text/plain charset=UTF-8' always;
          add_header 'Content-Length' 0 always;

          return 204;
        }

        add_header 'Access-Control-Allow-Origin' '*' always;
        add_header 'Access-Control-Allow-Credentials' 'true' always;
        add_header 'Access-Control-Allow-Methods' 'GET, POST, OPTIONS, DELETE, HEAD, PUT, PATCH' always;
        add_header 'Access-Control-Expose-Headers' 'DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Api-User,X-Api-Client,X-UA-Device-Category,X-UA-Device' always;
        add_header 'Access-Control-Allow-Headers' 'DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Api-User,X-Api-Client,X-UA-Device-Category,X-UA-Device' always;
        {{ else }}
           # this is a comment
        {{ end }}
    }

}
