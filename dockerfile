FROM ubuntu:16.04

MAINTAINER Grisha23

# Установка postgresql
RUN apt-get -y update
RUN apt-get -y install wget
RUN echo 'deb http://apt.postgresql.org/pub/repos/apt/ xenial-pgdg main' >> /etc/apt/sources.list.d/pgdg.list
RUN wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
RUN apt-get -y update
ENV PGVER 10
RUN apt-get -y install postgresql-$PGVER

# Adjust PostgreSQL configuration so that remote connections to the
# database are possible.
RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf

# And add ``listen_addresses`` to ``/etc/postgresql/$PGVER/main/postgresql.conf``
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf

USER root

# Установка golang
RUN apt install -y golang-1.10 git

# Выставляем переменную окружения для сборки проекта
ENV GOROOT /usr/lib/go-1.10
ENV GOPATH /opt/go
ENV PATH $GOROOT/bin:$GOPATH/bin:/usr/local/go/bin:$PATH


# Копируем исходный код в Docker-контейнер
WORKDIR $GOPATH/src/github.com/Grisha23/ForumsApi/
ADD ./ $GOPATH/src/github.com/Grisha23/ForumsApi/

# Подтягиваем зависимости
RUN go get \
    github.com/lib/pq \
    github.com/gorilla/mux \
    github.com/prometheus/client_golang/prometheus \
	github.com/prometheus/client_golang/prometheus/promhttp

# Собираем пакет
RUN go install .
EXPOSE 5000


# Run the rest of the commands as the ``postgres`` user created by the ``postgres-$PGVER`` package when it was ``apt-get installed``
USER postgres

# Create a PostgreSQL role named ``docker`` with ``docker`` as the password and
# then create a database `docker` owned by the ``docker`` role.
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker docker &&\
    psql -d docker &&\
    /etc/init.d/postgresql stop

# Expose the PostgreSQL port
EXPOSE 5000


CMD service postgresql start &&\
ForumsApi