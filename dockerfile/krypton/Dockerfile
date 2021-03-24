# Builder image
FROM golang:1.13-buster as builder

ARG KRYPTON_BRANCH=feature_adapt_k8s
RUN set -ex; \
    mkdir -p /go/src/github.com/andyli029; \
    cd /go/src/github.com/andyli029; \
    git clone --branch $KRYPTON_BRANCH https://github.com/andyli029/krypton.git; \
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
    apt-get install -y --no-install-recommends curl; \
    rm -rf /var/lib/apt/lists/*; \
    mkdir -p /krypton /etc/krypton /var/log/krypton /var/lib/krypton; \
    echo "/etc/krypton/krypton.json" > /krypton/config.path; \
    # allow to change config files
    chown -R mysql:mysql /krypton /etc/krypton /var/log/krypton /var/lib/krypton

VOLUME ["/var/lib/krypton", "/var/log/krypton"]

COPY --from=builder --chown=mysql:mysql /go/src/github.com/andyli029/krypton/bin/xenon /krypton/krypton
COPY --from=builder --chown=mysql:mysql /go/src/github.com/andyli029/krypton/bin/xenoncli /krypton/kryptoncli
COPY krypton-entry.sh /docker-entrypoint.sh
ENTRYPOINT ["/docker-entrypoint.sh"]

USER mysql
EXPOSE 8801
CMD ["/krypton/krypton", "-c", "/etc/krypton/krypton.json"]
