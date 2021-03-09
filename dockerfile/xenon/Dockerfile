# Builder image
FROM golang:1.13-buster as builder

ARG XENON_BRANCH=feature_adapt_k8s
RUN set -ex; \
    mkdir -p /go/src/github.com/andyli029; \
    cd /go/src/github.com/andyli029; \
    git clone --branch $XENON_BRANCH https://github.com/andyli029/krypton.git; \
    cd krypton; \
    make build

FROM ubuntu:focal

RUN set -ex; \
    groupadd --gid 999 --system mysql; \
    useradd \
    --uid 999 \
    --system \
    --home-dir /var/lib/mysql \
    --no-create-home \
    --gid mysql \
    mysql; \
    apt-get update; \
    apt-get install -y --no-install-recommends netcat; \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*; \
    mkdir -p /xenon /etc/xenon /var/log/xenon /var/lib/xenon; \
    echo "/etc/xenon/xenon.json" > /xenon/config.path; \
    # allow to change config files
    chown -R mysql:mysql /xenon /etc/xenon /var/log/xenon /var/lib/xenon

VOLUME ["/var/lib/xenon", "/var/log/xenon"]

COPY --from=builder --chown=mysql:mysql /go/src/github.com/andyli029/krypton/bin/xenon /xenon/xenon
COPY --from=builder --chown=mysql:mysql /go/src/github.com/andyli029/krypton/bin/xenoncli /xenon/xenoncli
COPY xenon-entry.sh /docker-entrypoint.sh
ENTRYPOINT ["/docker-entrypoint.sh"]

USER mysql
EXPOSE 8801
CMD ["/xenon/xenon", "-c", "/etc/xenon/xenon.json"]
