FROM oven/bun:1.2-alpine AS build

WORKDIR /app

COPY package.json bun.lock* ./
RUN bun install --production

COPY . .

FROM oven/bun:1.2-alpine AS runtime

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=build /app /app

USER appuser

EXPOSE 8080

CMD ["bun", "run", "src/index.ts"]
