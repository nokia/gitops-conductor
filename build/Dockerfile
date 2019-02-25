FROM gitopsbase

RUN mkdir /data && chown nobody:nobody /data
USER nobody

ADD build/_output/bin/gitops-conductor /usr/local/bin/gitops-conductor
