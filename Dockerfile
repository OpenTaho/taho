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
  tig \
  unzip \
  vim \
  wget \
  yamllint

RUN sh -c "$(wget -O- https://github.com/deluan/zsh-in-docker/releases/download/v1.2.1/zsh-in-docker.sh)"

RUN apt-get install -y sudo

RUN go install github.com/google/yamlfmt/cmd/yamlfmt@latest

RUN URL=https://api.github.com/repos/aquasecurity/tfsec/releases/latest \
  && URL="$(curl -s "$URL" | grep -o -E -m 1 "https://.+?tfsec-linux-amd64")" \
  && curl -s -L "$URL" > tfsec && chmod +x tfsec && mv tfsec /usr/bin

RUN URL=https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip \
  && curl -s "$URL" -o awscliv2.zip && unzip awscliv2.zip && ./aws/install \
  && rm -rf awscliv2.zip \
  && echo 'complete -C /usr/local/bin/aws_completer aws' >> /root/.zshrc

RUN git clone --depth 1 https://github.com/sheerun/vim-polyglot /root/.vim/pack/plugins/start/vim-polyglot

RUN wget -q https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 \
-O /usr/local/bin/yq \
&& chmod +x /usr/local/bin/yq

COPY .git /root/taho/.git

RUN cd /root/taho \
  && git reset --hard \
  && ./script install \
  && taho install-terraform \
  && taho install-terragrunt

RUN curl -s https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh \
  | bash

RUN cd /tmp || exit 1 \
  && curl -s -o packer.zip https://releases.hashicorp.com/packer/1.14.3/packer_1.14.3_linux_386.zip \
  && unzip packer.zip \
  && mv packer /usr/local/bin

COPY bin/export.sh /root
RUN cd /root \
  && echo 'source /root/export.sh' >> .zshrc \
  && PATH="/root/.tfenv/bin:/root/.tgenv/bin:$PATH"

RUN mkdir -p /root/bin

WORKDIR /workspace
