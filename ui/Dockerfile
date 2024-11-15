FROM node:22@sha256:5c76d05034644fa8ecc9c2aa84e0a83cd981d0ef13af5455b87b9adf5b216561 AS build
WORKDIR /app
COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile
COPY . .
ARG API_BASE_URL
ENV VITE_API_BASE_URL=$API_BASE_URL
RUN yarn build

FROM nginx:stable-alpine
WORKDIR /usr/share/nginx/html
COPY default.conf /etc/nginx/conf.d/default.conf
RUN rm -rf ./*
COPY --from=build /app/dist .
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
