FROM kapacitor:1.2

RUN apt-get update && \
    apt-get install -y curl

COPY entrypoint.sh /entrypoint.sh

EXPOSE 9092

ENTRYPOINT ["/entrypoint.sh"]
CMD ["kapacitord"]
