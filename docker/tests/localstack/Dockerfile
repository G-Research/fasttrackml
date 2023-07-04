FROM localstack/localstack
COPY --chown=localstack ./docker/tests/localstack/init-aws.sh /etc/localstack/init/ready.d/init-aws.sh
RUN chmod u+x /etc/localstack/init/ready.d/init-aws.sh