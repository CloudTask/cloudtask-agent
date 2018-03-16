FROM centos:7

MAINTAINER bobliu bobliu0909@gmail.com

RUN mkdir -p /opt/cloudtask/etc

RUN mkdir -p /opt/cloudtask/cache

RUN mkdir -p /opt/cloudtask/logs

COPY etc /opt/cloudtask/etc

COPY cloudtask-agnet /opt/cloudtask/cloudtask-agent

WORKDIR /opt/cloudtask

VOLUME ["/opt/cloudtask/etc"]

VOLUME ["/opt/cloudtask/cache"]

VOLUME ["/opt/cloudtask/logs"]

CMD ["./cloudtask-agent"]

EXPOSE 8600
