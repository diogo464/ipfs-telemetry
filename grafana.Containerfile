FROM docker.io/grafana/grafana:latest
COPY ./grafana/dashboards /var/lib/grafana/dashboards
COPY ./grafana/provisioning /etc/grafana/provisioning
ENV GF_SECURITY_ADMIN_PASSWORD="telemetry"
ENV GF_AUTH_ANONYMOUS_ENABLED="true"
