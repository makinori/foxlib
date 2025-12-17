# only run your servers in shell-less environments
# https://mastodon.hotmilk.space/@maki/115690185193300470

FROM ghcr.io/dart-musl/dart:latest AS dart

# ARG BUF_VERSION=1.61.0
# ARG SASS_VERSION=1.96.0

RUN \
# get buf
curl -Lo /usr/local/bin/buf \
# "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-$(uname -s)-$(uname -m)" && \
"https://github.com/bufbuild/buf/releases/latest/download/buf-$(uname -s)-$(uname -m)" && \
chmod +x /usr/local/bin/buf && \
# get dart-sass and compile
apk add --no-cache jq && \
SASS_VERSION=$(curl -s https://api.github.com/repos/sass/dart-sass/releases/latest | jq -r .tag_name) && \
apk del jq && \
git clone https://github.com/sass/dart-sass.git /dart-sass && \
cd /dart-sass && \
git checkout ${SASS_VERSION} && \
dart pub get && \
dart run grinder protobuf && \
dart compile exe bin/sass.dart -o /sass && \
# cleanup so we dont save all this
# apk seems to have written some files though
cd / && \
rm -f /usr/local/bin/buf && \
rm -rf /root/.cache /root/.dart-tool /root/.pub-cache /dart-sass

FROM scratch

WORKDIR /
ENV PATH=/:$PATH

COPY --from=dart /lib/ld-musl-*.so.1 /lib/
COPY --from=dart /sass /