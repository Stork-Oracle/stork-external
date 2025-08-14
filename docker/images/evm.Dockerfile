FROM node:20-alpine
COPY . /usr/src/app
WORKDIR /usr/src/app

ARG STORK_PUBLIC_KEY
ENV STORK_PUBLIC_KEY=${STORK_PUBLIC_KEY}

RUN npm install
RUN npm install -g wait-on

# Create the Hardhat config directory
RUN mkdir -p /root/.config/hardhat-nodejs

# Just run the entrypoint script directly - it already handles everything
ENTRYPOINT [ "/bin/sh", "-c", "/usr/src/app/docker-entrypoint.sh" ]
