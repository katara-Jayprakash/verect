# running ubuntu image on container
FROM ubuntu:focal
RUN apt-get update  -y
RUN apt-get install -y curl

# install node to container
RUN curl -sL https://deb.nodesource.com/setup_20.x | bash -
RUN apt-get upgrade -y
RUN apt-get install nodejs

# Getting running git on container
RUN apt-get install git -y

WORKDIR /home/app
COPY main.sh main.sh
ENTRYPOINT [ "/home/app/main.sh" ]
