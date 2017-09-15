FROM phusion/baseimage

RUN apt update && apt upgrade -y -o Dpkg::Options::="--force-confold"
RUN apt install -y fakeroot git rpm cpio zip python-pip

RUN curl -L https://github.com/hogliux/bomutils/tarball/master | tar xz && cd hogliux-bomutils-* && make install

RUN apt install -y libxml2-dev libssl-dev pkg-config
RUN curl -L https://github.com/downloads/mackyle/xar/xar-1.6.1.tar.gz | tar xz && cd xar* && ./configure && make && make install

RUN pip install awscli

# for debian repository generation
RUN apt install -y ruby1.9.1 createrepo
RUN gem install deb-s3

# for rpmsigning process
RUN apt install -y expect

# osslsigncode
RUN apt install -y autoconf build-essential libcurl4-openssl-dev

RUN cd /tmp && \
  curl -L https://downloads.sourceforge.net/project/osslsigncode/osslsigncode/osslsigncode-1.7.1.tar.gz | \
    tar xzf - && \
  cd osslsigncode-1.7.1 && \
  ./configure && \
  make && \
  make install && \
  cd .. && \
  rm -rf osslsigncode-1.7.1
