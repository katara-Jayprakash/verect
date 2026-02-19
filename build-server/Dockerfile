# Use Ubuntu base
FROM ubuntu:focal
RUN apt-get update  -y

# installing curl and git
RUN apt-get install -y curl
RUN apt-get install git -y


# install node to container
RUN curl -sL https://deb.nodesource.com/setup_20.x | bash -
RUN apt-get upgrade -y
RUN apt-get install nodejs -y


WORKDIR /home/app
COPY main.sh main.sh
COPY script.js script.js
COPY package.json .
COPY pnpm*.yaml .

# Install pnpm globally
RUN npm install -g pnpm
RUN pnpm install

# giving commands their rights
RUN chmod +x main.sh
RUN chmod +x script.js

ENTRYPOINT [ "/home/app/main.sh" ]
