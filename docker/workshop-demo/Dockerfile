FROM node:alpine AS builder

WORKDIR /app

COPY ./web/demo/package*.json ./

RUN npm install

COPY ./web/demo/. .

ARG VITE_WORKSHOP_HOST
ENV VITE_WORKSHOP_HOST=$VITE_WORKSHOP_HOST

RUN npm run build

FROM nginx:alpine AS deploy

COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
