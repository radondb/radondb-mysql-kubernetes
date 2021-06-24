FROM ubuntu:focal

RUN set -ex; \
    groupadd --gid 999 --system mysql; \
    useradd \
    --uid 999 \
    --system \
    --home-dir /var/lib/mysql \
    --no-create-home \
    --gid mysql \
    mysql

ENV PS_VERSION 5.7.34-37-1
ENV OS_VER focal
ENV FULL_PERCONA_VERSION "$PS_VERSION.$OS_VER"

RUN set -ex; \
    apt-get update; \
    apt-get install -y --no-install-recommends gnupg2 wget lsb-release curl; \
    wget -P /tmp --no-check-certificate https://repo.percona.com/apt/percona-release_latest.$(lsb_release -sc)_all.deb; \
    dpkg -i /tmp/percona-release_latest.$(lsb_release -sc)_all.deb; \
    apt-get update; \
    export DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true; \
    { \
        echo percona-server-server-5.7=${FULL_PERCONA_VERSION} percona-server-server-5.7=${FULL_PERCONA_VERSION}/root-pass password ''; \
        echo percona-server-server-5.7=${FULL_PERCONA_VERSION} percona-server-server-5.7=${FULL_PERCONA_VERSION}/re-root-pass password ''; \
        echo tzdata tzdata/Areas select Asia; \
        echo tzdata tzdata/Zones/Asia select Shanghai; \
    } | debconf-set-selections; \
    # install "tzdata" for /usr/share/zoneinfo/
    apt-get install -y --no-install-recommends libjemalloc1 libmecab2 tzdata; \
    apt-get install -y --no-install-recommends \
        percona-server-server-5.7=${FULL_PERCONA_VERSION} \
        percona-server-common-5.7=${FULL_PERCONA_VERSION} \
        percona-server-tokudb-5.7=${FULL_PERCONA_VERSION}; \
    # TokuDB modifications
    echo "LD_PRELOAD=/usr/lib64/libjemalloc.so.1" >> /etc/default/mysql; \
    echo "THP_SETTING=never" >> /etc/default/mysql; \
    \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* /var/lib/mysql /etc/alternatives/my.cnf /etc/mysql/*; \
    mkdir -p /var/lib/mysql /var/log/mysql /var/run/mysqld /etc/mysql/conf.d /docker-entrypoint-initdb.d; \
    # allow to change config files
    chown -R mysql:mysql /var/lib/mysql /var/log/mysql /var/run/mysqld /etc/mysql; \
    # ensure that /var/run/mysqld (used for socket and lock files) is writable regardless of the UID our mysqld instance ends up having at runtime
    chmod 1777 /var/run/mysqld

VOLUME ["/var/lib/mysql", "/var/log/mysql"]

COPY mysql-entry.sh /docker-entrypoint.sh
COPY --chown=mysql:mysql my.cnf /etc/mysql/my.cnf
ENTRYPOINT ["/docker-entrypoint.sh"]

EXPOSE 3306
CMD ["mysqld"]
