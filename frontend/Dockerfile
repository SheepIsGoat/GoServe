# Step 1: Install necessary packages
FROM node:14 AS install
WORKDIR /app
COPY package*.json ./
RUN npm install

# Step 2: Build the React Application
FROM install AS build
COPY . ./
RUN npm run build

# Step 3: Serve the React application
FROM nginx:alpine

RUN apk add --no-cache openssl
COPY nginx.conf.template /etc/nginx/nginx.conf.template
COPY --from=build /app/build /usr/share/nginx/html
# COPY default.conf /etc/nginx/conf.d/default.conf
EXPOSE 80 443

COPY entrypoint.sh /
RUN chmod +x entrypoint.sh
ENTRYPOINT [ "/entrypoint.sh" ]

CMD ["nginx", "-g", "daemon off;"]