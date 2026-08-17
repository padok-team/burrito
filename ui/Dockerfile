FROM ghcr.io/pnpm/pnpm:11.21.0@sha256:ff1b4619e8299734f35afb21bd2be18f51b7b6e49cd30fb14300fadeafe871da AS build
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY . .
ARG API_BASE_URL
ENV VITE_API_BASE_URL=$API_BASE_URL
RUN pnpm run build

FROM nginx:stable-alpine@sha256:97d490c12ba55b4946b01546d1c3ed324e8d41ab1c9fcb2a616aa470620e5b46
WORKDIR /usr/share/nginx/html
COPY default.conf /etc/nginx/conf.d/default.conf
RUN rm -rf ./*
COPY --from=build /app/dist .
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
