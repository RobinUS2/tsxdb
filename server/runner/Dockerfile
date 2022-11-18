FROM ubuntu:latest
MAINTAINER Robin Verlangen

# Update
RUN apt-get update && \
    apt-get dist-upgrade -y

# Debug tools
RUN apt-get update && \
    apt-get install dstat telnet redis-tools htop -y

# User (+group) to run as non-root
RUN groupadd -r tsxdb && useradd -r -g tsxdb tsxdb

# Copy binary
RUN mkdir -p /usr/local/tsxdb/bin
COPY --chown=tsxdb:tsxdb runner /usr/local/tsxdb/bin/tsxdb
RUN chmod +x /usr/local/tsxdb/bin/tsxdb

# Config
RUN mkdir -p /etc/tsxdb
COPY --chown=tsxdb:tsxdb config.yaml /etc/tsxdb/

# Fix folder permissions
RUN chown -R tsxdb:tsxdb /usr/local/tsxdb /etc/tsxdb

# Non-root
USER tsxdb:tsxdb

# Start
CMD ["/usr/local/tsxdb/bin/tsxdb"]