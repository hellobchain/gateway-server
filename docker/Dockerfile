FROM golang:1.25 as build
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /gateway-server
COPY . .
RUN rm -rf go.sum && make
WORKDIR /build
RUN cp -r /gateway-server/bin .
RUN cp config.yml .
FROM alpine:3.14
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories \
    && apk update \
    && apk add tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone
WORKDIR /gateway-server
COPY --from=build /build/config.yml /gateway-server/config.yml
COPY --from=build /build/bin/gateway-server.bin /usr/local/bin
WORKDIR /gateway-server
CMD gateway-server.bin /gateway-server/config.yml





