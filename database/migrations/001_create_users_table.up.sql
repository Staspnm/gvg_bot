CREATE TABLE users (
id SERIAL PRIMARY KEY,
telegram_id BIGINT UNIQUE NOT NULL,
game_nickname VARCHAR(100) NOT NULL,
nine_digit_code VARCHAR(9) UNIQUE NOT NULL,
guild_name VARCHAR(100) NOT NULL,
guild_role VARCHAR(20) NOT NULL CHECK (guild_role IN ('owner', 'leader', 'officer', 'member')),
is_active BOOLEAN DEFAULT TRUE,
registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_users_nine_digit_code ON users (nine_digit_code);