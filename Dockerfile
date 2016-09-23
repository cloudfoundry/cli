FROM golang

RUN go get golang.org/x/tools/cmd/cover

RUN sed -i -e 's/httpredir.debian.org/ftp.us.debian.org/' /etc/apt/sources.list
RUN apt-get update && apt-get -y install fakeroot
RUN apt-get update && apt-get -y install rpm

RUN curl -L https://github.com/hogliux/bomutils/tarball/master | tar xz && cd hogliux-bomutils-* && make install

RUN apt-get update && apt-get -y install libxml2-dev libssl-dev
RUN curl -L https://github.com/downloads/mackyle/xar/xar-1.6.1.tar.gz | tar xz && cd xar* && ./configure && make && make install

RUN apt-get update && apt-get -y install cpio

RUN apt-get update && apt-get -y install zip

RUN apt-get update && apt-get -y install python-pip
RUN pip install awscli

RUN apt-get update && apt-get -y install jq

# for debian repository generation
RUN apt-get update && apt-get -y install ruby1.9.1
RUN gem install deb-s3
RUN apt-get update && apt-get -y install createrepo

# for rpmsigning process
RUN apt-get update && apt-get -y install expect

# osslsigncode
RUN apt-get -y update && \
  apt-get -y install \
    autoconf \
    build-essential \
    libcurl4-openssl-dev \
    libssl-dev

RUN cd /tmp && \
  curl -L https://downloads.sourceforge.net/project/osslsigncode/osslsigncode/osslsigncode-1.7.1.tar.gz | \
    tar xzf - && \
  cd osslsigncode-1.7.1 && \
  ./configure && \
  make && \
  make install && \
  cd .. && \
  rm -rf osslsigncode-1.7.1
