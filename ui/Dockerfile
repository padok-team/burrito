FROM node:24@sha256:00e9195ebd49985a6da8921f419978d85dfe354589755192dc090425ce4da2f7 AS build
WORKDIR /app
COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile
COPY . .
ARG API_BASE_URL
ENV VITE_API_BASE_URL=$API_BASE_URL
RUN yarn build

FROM nginx:stable-alpine@sha256:15e96e59aa3b0aada3a121296e3bce117721f42d88f5f64217ef4b18f458c6ab
WORKDIR /usr/share/nginx/html
COPY default.conf /etc/nginx/conf.d/default.conf
RUN rm -rf ./*
COPY --from=build /app/dist .
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
