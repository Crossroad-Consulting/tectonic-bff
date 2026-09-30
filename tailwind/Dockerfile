FROM node:26-bookworm-slim

WORKDIR /app

COPY package.json package-lock.json ./

RUN npm ci --no-audit --no-fund

CMD ["npx", "@tailwindcss/cli", "-i", "/app/src/styles.css", "-o", "/app/dst/styles.css", "--watch"]