FROM node:18-alpine AS frontend

WORKDIR /build

COPY ./package.json ./package.json
COPY ./package-lock.json ./package-lock.json

RUN --mount=type=cache,target=node_modules npm ci

COPY ./.config ./.config
COPY ./webpack.config.ts ./webpack.config.ts
COPY ./tsconfig.json ./tsconfig.json
COPY ./.eslintrc ./.eslintrc

COPY ./src ./src
# FIXME: See #38.
# RUN --mount=type=cache,target=node_modules npm run test

COPY ./README.md ./README.md
COPY ./CHANGELOG.md ./CHANGELOG.md
COPY ./LICENSE ./LICENSE
RUN --mount=type=cache,target=node_modules npm run build

FROM scratch

COPY --from=frontend /build/dist /
