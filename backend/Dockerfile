# Stage 1: Build Dependencies
FROM node:latest AS build

WORKDIR /usr/src/app

COPY package*.json ./

RUN npm install && \
    npm install -g typescript ts-node

# Stage 2: Copy Local Code and Install Global Dependencies
FROM node:latest AS app

WORKDIR /usr/src/app

COPY --from=build /usr/src/app/node_modules ./node_modules
COPY . ./

RUN npm install -g typescript ts-node && \
    tsc

# Stage 3: Copy Compiled server.js
FROM node:latest AS final
WORKDIR /usr/src/app
COPY --from=app /usr/src/app/dist/server.js ./server.js
COPY --from=app /usr/src/app/dist/server ./server
COPY --from=build /usr/src/app/node_modules ./node_modules

EXPOSE 3000

CMD ["node", "server.js"]