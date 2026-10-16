# Development image: runs the Vite dev server. src/ and index.html are
# bind-mounted by compose.yaml for hot reload.
FROM node:24-alpine
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
EXPOSE 5173
CMD ["npm", "run", "dev"]
