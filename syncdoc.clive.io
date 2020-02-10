upstream syncdoc_clive_io_upstream {
  server 127.0.0.1:3001;
  keepalive 8; # will this prevent more than 8 users simultaneously?
}

server {
    listen 443 http2 ssl;
    listen [::]:443 http2 ssl;

    server_name syncdoc.clive.io;
    root /home/clive/code/syncdoc/static;

    sendfile on;

    location = / {
    }

    location ~ ^/[a-zA-Z0-9\-\_]+$ {
        try_files /app.html /app.html;
    }

    error_page 404 /404.html;
    location = /404.html {
        internal;
    }

    location /ws/ {
        proxy_pass http://syncdoc_clive_io_upstream;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
    }

    #client_max_body_size 100k; # or does this interfere with long ws sessions?
    #client_body_timeout 120s;
}
