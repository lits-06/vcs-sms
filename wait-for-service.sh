#!/bin/sh
set -e
for service in $WAIT_FOR_SERVICES; do
  host=$(echo $service | cut -d':' -f1)
  port=$(echo $service | cut -d':' -f2)
  until nc -z "$host" "$port"; do
    echo "Waiting for $host:$port..."
    sleep 5
  done
done

# Khi tất cả service sẵn sàng, chạy lệnh chính
exec $COMMAND
