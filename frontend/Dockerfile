FROM node:22-alpine AS build
WORKDIR /app
COPY package.json ./
RUN npm install
COPY . .
RUN npx vite build

FROM nginx:1.27-alpine
RUN apk add --no-cache openssl \
  && openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
    -keyout /etc/nginx/gate.key -out /etc/nginx/gate.crt \
    -subj "/CN=the-gate" \
  && chown -R nginx:nginx /etc/nginx/gate.key /etc/nginx/gate.crt
COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8443
USER nginx
CMD ["nginx", "-g", "daemon off;"]