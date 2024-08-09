-- Creat the Host Information table
CREATE TABLE IF NOT EXISTS agents (
    host_id SERIAL PRIMARY KEY,
    hostname VARCHAR(255) NOT NULL,
    ip_address VARCHAR(15),
    os VARCHAR(50),
    os_version VARCHAR(50),
    hardware_specs JSONB,
    agent_version VARCHAR(50),
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    last_user VARCHAR(255)
);

-- Create the device_groups Table
CREATE TABLE IF NOT EXISTS device_groups (
    group_id SERIAL PRIMARY KEY,
    group_name VARCHAR(255) UNIQUE NOT NULL
);

-- Create the device_group_members Table
CREATE TABLE IF NOT EXISTS device_group_members (
    host_id INT NOT NULL,
    group_id INT NOT NULL,
    PRIMARY KEY (host_id, group_id),
    FOREIGN KEY (host_id) REFERENCES agents(host_id),
    FOREIGN KEY (group_id) REFERENCES device_groups(group_id)
);

-- Create the User Table
CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL
);