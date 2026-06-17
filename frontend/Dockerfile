FROM node:22-alpine AS build

WORKDIR /app

ARG VITE_API_BASE_URL=http://localhost:8080
ARG VITE_WS_BASE_URL=ws://localhost:8080/ws
ARG VITE_STUDENT_BASE_URL=http://localhost:3000/student

ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
ENV VITE_WS_BASE_URL=$VITE_WS_BASE_URL
ENV VITE_STUDENT_BASE_URL=$VITE_STUDENT_BASE_URL

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM node:22-alpine

WORKDIR /app

ENV PORT=4173

COPY --from=build /app/dist ./dist
COPY docker-server.mjs ./docker-server.mjs

EXPOSE 4173

CMD ["node", "docker-server.mjs"]
