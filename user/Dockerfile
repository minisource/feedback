# Base image
FROM node:18-alpine AS base

ARG PORT=3001

# Dependencies
FROM base AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

# Builder
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build

# Production runner
FROM base AS runner
WORKDIR /app
ENV PORT=$PORT
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs
COPY --from=builder --chown=nextjs:nodejs /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static
COPY --chown=nextjs:nodejs .env ./
USER nextjs
EXPOSE $PORT

# Option 1: اگر server.js وجود داره
# CMD ["node", "server.js"]

# Option 2: اگر Next.js 15 دیگه server.js نسازه
CMD ["sh", "-c", "export $(cat .env | xargs) && npx next start -H 0.0.0.0"]
