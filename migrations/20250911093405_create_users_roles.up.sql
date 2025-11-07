CREATE TABLE user_roles
(
    id      SERIAL PRIMARY KEY,
    user_id UUID        NOT NULL,
    role    VARCHAR(50) NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id)