#!/bin/bash

set -e
set -u

function create_user_and_database() {
	local database=$1
	echo "  Creating user and database '$database'"
	psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
	    CREATE USER ${database};
	    CREATE DATABASE ${database};
	    GRANT ALL PRIVILEGES ON DATABASE ${database} TO ${database};

			\c ${database};

			CREATE TABLE IF NOT EXISTS idempotency_keys (
				id uuid PRIMARY KEY
			);

			DO \$$ BEGIN
				CREATE TYPE notification_status AS ENUM ('SENT', 'FAILED', 'PENDING');
			EXCEPTION
				WHEN duplicate_object THEN null;
			END \$$;


			DO \$$ BEGIN
				CREATE TYPE notification_type AS ENUM ('SMS', 'SLACK', 'EMAIL');
			EXCEPTION
				WHEN duplicate_object THEN null;
			END \$$;

			CREATE TABLE IF NOT EXISTS notifications (
				id 				 uuid 						 	 				PRIMARY KEY,
				type 			 notification_type 	 				NOT NULL,
				status 		 notification_status 				NOT NULL,
				data 			 jsonb 							 				NOT NULL,
				created_at timestamp without time zone NOT NULL,
				updated_at timestamp without time zone NOT NULL
			);
EOSQL
}

if [[ -n "$POSTGRES_MULTIPLE_DATABASES" ]]; then
	echo "Multiple database creation requested: $POSTGRES_MULTIPLE_DATABASES"
	for database in $(echo "${POSTGRES_MULTIPLE_DATABASES}" | tr ',' ' '); do
		create_user_and_database "${database}"
	done
	echo "Multiple databases created"
fi
