FROM node:20-alpine AS build
WORKDIR /app
COPY package.json ./
RUN corepack enable && pnpm i
COPY . .
RUN pnpm build

FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
