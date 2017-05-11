# DKCONF a golang tool to templatize config file with environment variables

[![Build Status](https://travis-ci.org/mikrob/dkconf.svg?branch=master)](https://travis-ci.org/mikrob/dkconf)

Often when we make docker images, we would like to pass config from environment variables to config files.
But there were no tools to do that.

DkConf do it for you!

## Usage

```bash
#> dkconf -h
Usage of ./dkconf-osx:
  -p string
    	env var prefix (default "APPCONF")
  -s string
    	absolute path to the source template file
  -t string
    	absolute path to the target file generated
```

dkconf as two mode :

1. Write processed template directly to target file

```bash
dkconf -s ./examples/nginx-vhost.conf.tpl -t output_example/nginx_vhost.conf -p NGX
```

2. Dump processed template to stdout

```bash
dkconf -s ./examples/nginx-vhost.conf.tpl -p NGX
```

-p parameters definie the environment variable prefix used.

## Variable format

Template language used is go template, good tutorial here [https://gohugo.io/templates/go-templates/](https://gohugo.io/templates/go-templates/)

### Commons

In template you should use camelCase name such as : `MyVarIsInCamelCaseFormat`

The corresponding env var will be in bash style such as : `MY_VAR_IS_IN_CAMEL_CASE_FORMAT`

### boolean

You can use boolean strictly.
Example :

`export NGX_CORS_ENABLED=true`
`export NGX_CORS_ENABLED=false`

Only true or false will be accepted to make test like this in templates :

```golang
{{ if .CorsEnabled }}
...
{{ else }}
...
{{ end }}
```

### Lists

You canse use list in your variables, a list is determined by elements separted by a comma : `,`

Example :

```bash
export NGX_SECONDARY_SERVERS_NAMES="tata.com,tutu.com,site1.toto.com,site3.sub.toto.com,test.pouet.com"
```

In your template you will be able to iterate over the list in this way :

```golang
{{ range $idx, $elem := .SecondaryServersNames}}
  {{ $elem }}
{{ end }}
```

### Undefined variables

if you declare a variable in your template which is not available as environment variable DkConf will put a message in the generated template such as :

```bash
####### DKCONF : MISSING ENV VAR FOR GO TPL VALUE: TrucBidule, SHOULD BE NGX_TRUC_BIDULE #######
```

Here we had declared `{{ .TrucBidule }}` in the template

## Example

Let's admit you make a docker image with nginx.
It would be great if you could for example specify multiples server names, and options to enable cors or not on environment variables.

So for example :

```bash
export NGX_FQDN=tutu.com
export NGX_CORS_ENABLED=true
export NGX_MAIN_SERVER_NAME="tutu.com"
export NGX_SECONDARY_SERVERS_NAMES="tata.com,tutu.com,site1.toto.com,site3.sub.toto.com,test.pouet.com"
```


Let's make a nginx vhost template :

```nginx
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
```

Let's process it  :

```bash
dkconf -s ./examples/nginx-vhost.conf.tpl -p NGX
```

Output will be :

```bash
server {
    listen      80;
    root        /app/web;

    server_name tata.com tutu.com site1.toto.com site3.sub.toto.com test.pouet.com ;
    access_log  /var/log/nginx/access.log main;
    error_log   /var/log/nginx/error.log;

    ####### DKCONF : MISSING ENV VAR FOR GO TPL VALUE: TrucBidule, SHOULD BE NGX_TRUC_BIDULE #######

    tutu.com

    location / {
        if ($http_x_forwarded_proto != "https") {
          return 301 https://www.tutu.com$request_uri;
        }
        if ($host != "www.tutu.com") {
          return 301 https://www.tutu.com$request_uri;
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

    }

}
```
