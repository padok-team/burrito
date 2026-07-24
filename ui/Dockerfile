FROM node:24@sha256:5711a0d445a1af54af9589066c646df387d1831a608226f4cd694fc59e745059 AS build
WORKDIR /app
COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile
COPY . .
ARG API_BASE_URL
ENV VITE_API_BASE_URL=$API_BASE_URL
RUN yarn build

FROM nginx:stable-alpine@sha256:97d490c12ba55b4946b01546d1c3ed324e8d41ab1c9fcb2a616aa470620e5b46
WORKDIR /usr/share/nginx/html
COPY default.conf /etc/nginx/conf.d/default.conf
RUN rm -rf ./*
COPY --from=build /app/dist .
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
