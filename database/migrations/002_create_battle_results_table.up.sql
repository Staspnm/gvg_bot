CREATE TABLE battle_results (
id SERIAL PRIMARY KEY,
user_id INTEGER REFERENCES users(id),
location VARCHAR(2) NOT NULL CHECK (location IN ('T1', 'T2', 'T3', 'T4', 'B1', 'B2', 'B3', 'B4', 'F1', 'F2')),
enemy_squad VARCHAR(100) NOT NULL,
own_squad VARCHAR(100) NOT NULL,
flags_count INTEGER NOT NULL CHECK (flags_count BETWEEN 1 AND 22),
reported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
guild_name VARCHAR(100) NOT NULL,
battle_date DATE NOT NULL DEFAULT CURRENT_DATE
);

CREATE INDEX idx_battle_results_guild_location_date ON battle_results (guild_name, location, battle_date);