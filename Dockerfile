FROM ubuntu:latest

RUN yes | unminimize

# Install tools using apt-get
RUN apt-get update \
  && apt-get install -y \
  curl \
  dnsutils \
  git \
  graphviz \
  jq \
  man \
  markdownlint \
  python3-pip \
  python3.12-venv \
  tig \
  unzip \
  vim \
  wget \
  yamllint

RUN sh -c "$( \
  wget -O- https://github.com/deluan/zsh-in-docker/releases/download/v1.2.1/zsh-in-docker.sh \
  )"

RUN apt-get install -y sudo


RUN mkdir /root/Downloads \
  && cd /root/Downloads \
  && arch="$(arch | sed 's/aarch64/arm64/')" \
  && url="https://go.dev/dl/go1.26.0.linux-$arch.tar.gz" \
  && curl -sL "$url" -o go.tar.gz \
  && tar -C /usr/local -xzf go.tar.gz \
  && PATH="$PATH:/usr/local/go/bin" \
  && go install github.com/google/yamlfmt/cmd/yamlfmt@latest

RUN URL='https://api.github.com/repos/aquasecurity/tfsec/releases/latest' \
  && URL="$(curl -s "$URL" | grep -o -E -m 1 "https://.+?tfsec-linux-amd64")" \
  && curl -s -L "$URL" > tfsec && chmod +x tfsec && mv tfsec /usr/bin

RUN URL="https://awscli.amazonaws.com/awscli-exe-linux-$(arch).zip" \
  && curl -s "$URL" -o awscliv2.zip \
  && unzip -q awscliv2.zip \
  && ./aws/install \
  && rm -rf awscliv2.zip \
  && echo 'complete -C /usr/local/bin/aws_completer aws' >> /root/.zshrc

RUN git clone --depth 1 \
  'https://github.com/sheerun/vim-polyglot' \
  /root/.vim/pack/plugins/start/vim-polyglot

RUN wget -q 'https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64' \
-O /usr/local/bin/yq \
&& chmod +x /usr/local/bin/yq

COPY .git /root/taho/.git

RUN cd /root/taho \
  && git reset --hard \
  && ./script install \
  && taho install-k \
  && taho install-terraform \
  && taho install-terragrunt \
  && echo 'source <(/root/bin/kubectl completion zsh)' >> /root/.zshrc

RUN curl -s 'https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh' \
  | bash

RUN cd /tmp || exit 1 \
  && curl -s -o packer.zip 'https://releases.hashicorp.com/packer/1.14.3/packer_1.14.3_linux_arm64.zip' \
  && unzip packer.zip \
  && mv packer /usr/local/bin

COPY bin/export.sh /root
RUN cd /root \
  && echo 'source /root/export.sh' >> .zshrc \
  && PATH="/root/.tfenv/bin:/root/.tgenv/bin:$PATH"

RUN cd /root/bin \
  && curl -sLO 'https://raw.githubusercontent.com/ahmetb/kubectx/refs/heads/master/kubens' \
  && chmod +x 'kubens' \
  && curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" \
  | bash

RUN curl -fsSL -o get_helm.sh 'https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-4' \
  && chmod 700 get_helm.sh \
  && ./get_helm.sh

RUN curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" \
  | bash

RUN PLATFORM="$(uname -s)_arm64" \
  && curl -sLO "https://github.com/eksctl-io/eksctl/releases/latest/download/eksctl_$PLATFORM.tar.gz" \
  && tar -xzf eksctl_$PLATFORM.tar.gz -C /tmp && rm eksctl_$PLATFORM.tar.gz \
  && install -m 0755 /tmp/eksctl /usr/local/bin \
  && rm /tmp/eksctl

WORKDIR /workspace
