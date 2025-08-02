# Copyright (C) 2023-2025  Eric Cornelissen
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.

ARG GO_VERSION=invalid
FROM docker.io/golang:${GO_VERSION} AS build

WORKDIR /src

COPY cmd/ ./cmd/
COPY go.mod go.sum *.go ./

RUN go run tasks.go build

# ---

FROM scratch AS main

LABEL org.opencontainers.image.title="ades" \
	org.opencontainers.image.description="Find dangerous uses of GitHub Actions Workflow expressions." \
	org.opencontainers.image.version="v25.07" \
	org.opencontainers.image.licenses="GPL-3.0-or-later" \
	org.opencontainers.image.source="https://github.com/ericcornelissen/ades"

COPY --from=build /src/ades /bin/ades
COPY COPYING.txt SECURITY.md /

WORKDIR /src
VOLUME /src

USER 1000:1000
USER 1000

ENTRYPOINT ["/bin/ades"]
