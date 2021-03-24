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
    if ! which gpg; then \
        apt-get install -y --no-install-recommends gnupg; \
    fi; \
    if ! gpg --version | grep -q '^gpg (GnuPG) 1\.'; then \
        apt-get install -y --no-install-recommends dirmngr; \
    fi; \
    rm -rf /var/lib/apt/lists/*

RUN set -ex; \
    key='9334A25F8507EFA5'; \
    export GNUPGHOME="$(mktemp -d)"; \
    gpg --batch --keyserver keys.gnupg.net --recv-keys "$key"; \
    gpg --batch --export $key > /etc/apt/trusted.gpg.d/mysql.gpg; \
    command -v gpgconf > /dev/null && gpgconf --kill all || :; \
    rm -r "$GNUPGHOME"; \
    apt-key list > /dev/null; \
    echo "deb http://repo.percona.com/apt focal main\ndeb-src http://repo.percona.com/apt focal main">> /etc/apt/sources.list.d/mysql.list; \
    apt-get update; \
    export DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true; \
    { \
        echo percona-server-server-5.7 percona-server-server-5.7/root-pass password ''; \
        echo percona-server-server-5.7 percona-server-server-5.7/re-root-pass password ''; \
        echo tzdata tzdata/Areas select Asia; \
        echo tzdata tzdata/Zones/Asia select Shanghai; \
    } | debconf-set-selections; \
    # install "tzdata" for /usr/share/zoneinfo/
    apt-get install -y --no-install-recommends libjemalloc1 libmecab2 tzdata; \
    apt-get install -y --no-install-recommends \
        percona-server-server-5.7 \
        percona-server-common-5.7 \
        percona-server-tokudb-5.7; \
    # TokuDB modifications
    echo "LD_PRELOAD=/usr/lib64/libjemalloc.so.1" >> /etc/default/mysql; \
    echo "THP_SETTING=never" >> /etc/default/mysql; \
    \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* /var/lib/mysql /etc/alternatives/my.cnf /etc/mysql/*; \
    mkdir -p /var/lib/mysql /var/log/mysql /var/run/mysqld /etc/mysql/conf.d; \
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