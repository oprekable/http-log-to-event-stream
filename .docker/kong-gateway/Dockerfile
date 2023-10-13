FROM kong:2.4

COPY ./.docker/kong-gateway/kong.conf /etc/kong/kong.conf
COPY ./.docker/kong-gateway/declarative /home/kong
COPY ./.docker/kong-gateway/plugins /usr/local/share/lua/5.1/kong/plugins

CMD [ "kong", "start" ]
