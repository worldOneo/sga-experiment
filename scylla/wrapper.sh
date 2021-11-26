#!/bin/bash

CQL="CREATE KEYSPACE IF NOT EXISTS todos
		WITH REPLICATION = { 
			'class' : 'SimpleStrategy', 
			'replication_factor' : 1 
		 };"
echo "Executing: $CQL"

until cqlsh -e "$CQL"; do
    echo "Unavailable: sleeping"
    sleep 10
done &

exec /docker-entrypoint.py "$@"
