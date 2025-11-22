#!/bin/bash

CONTAINER_NAME="peril_rabbitmq"
IMAGE_NAME="rabbitmq-stomp"

start_or_run () {
    docker inspect $CONTAINER_NAME > /dev/null 2>&1

    if [ $? -eq 0 ]; then
        echo "Starting Peril RabbitMQ container..."
        docker start $CONTAINER_NAME
    else
        echo "Peril RabbitMQ container not found, creating a new one..."
        docker run -d --name $CONTAINER_NAME -p 5672:5672 -p 15672:15672 -p 61613:61613 $IMAGE_NAME
    fi
}

case "$1" in
    start)
        start_or_run
        ;;
    stop)
        echo "Stopping Peril RabbitMQ container..."
        docker stop $CONTAINER_NAME
        ;;
    logs)
        echo "Fetching logs for Peril RabbitMQ container..."
        docker logs -f $CONTAINER_NAME
        ;;
    *)
        echo "Usage: $0 {start|stop|logs}"
        exit 1
esac
