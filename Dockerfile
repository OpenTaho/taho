FROM ubuntu:latest

# Install tools using apt-get
RUN apt-get update \
  && apt-get install -y \
  curl \
  dnsutils \
  git \
  golang \
  jq \
  markdownlint \
  python3-pip \
  unzip \
  vim \
  yamllint

RUN go install github.com/google/yamlfmt/cmd/yamlfmt@latest

RUN URL=https://raw.githubusercontent.com/terraform-linters/tflint \
  && URL="$URL/master/install_linux.sh" \
  && curl -s "$URL" | bash

RUN URL=https://api.github.com/repos/aquasecurity/tfsec/releases/latest \
  && URL="$(curl -s "$URL" | grep -o -E -m 1 "https://.+?tfsec-linux-amd64")" \
  && curl -s -L "$URL" > tfsec && chmod +x tfsec && mv tfsec /usr/bin

RUN URL=https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip \
  && curl -s "$URL" -o awscliv2.zip && unzip awscliv2.zip && ./aws/install \
  && rm -rf awscliv2.zip

RUN git clone --depth 1 https://github.com/tfutils/tfenv.git "$HOME/.tfenv"
RUN git clone --depth 1 https://github.com/tgenv/tgenv.git "$HOME/.tgenv"

COPY .git /root/taho/.git

RUN cd /root/taho \
  && git reset --hard \
  && ./script install

COPY bin/export.sh bin/aliases.sh /root/
RUN cd /root \
  && cat /root/export.sh >> .profile \
  && cat /root/aliases.sh >> .bashrc \
  && echo 'source /root/export.sh' >> .bashrc \
  && PATH="/root/.tfenv/bin:/root/.tgenv/bin:$PATH" \
  && tgenv install 0.83.1 \
  && tfenv use latest \
  && tgenv use latest

RUN mkdir -p /root/bin \
  && mkdir -p /workspace \
  && ln -s /workspace/bin/tf-civl.sh /usr/bin/tf-civl \
  && ln -s /workspace/bin/tf-modules-upgrade.sh /usr/bin/tf-modules-upgrade \
  && ln -s /workspace/bin/tgf-rm-cache-dirs.sh /usr/bin/tgf-rm-cache-dirs

WORKDIR /workspace
