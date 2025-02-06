-- Create Zabbix database
CREATE DATABASE zabbix;

-- Create Zabbix schema (will be populated by Zabbix server on first run)
\c zabbix;

-- Set up permissions
GRANT ALL PRIVILEGES ON DATABASE zabbix TO ${POSTGRES_USER};
