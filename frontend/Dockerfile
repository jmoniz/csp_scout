# Stage 1: Build SvelteKit application
FROM node:22-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

# Stage 2: Minimal production Node.js runner
FROM node:22-alpine AS runner

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci --omit=dev

COPY --from=builder /app/build ./build

ENV NODE_ENV=production
ENV PORT=3000
ENV BACKEND_INTERNAL_URL=http://backend:8080

EXPOSE 3000

CMD ["node", "build"]
