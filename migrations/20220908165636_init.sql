-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users
(
    id SERIAL PRIMARY KEY,
    name VARCHAR(30) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS events
(
    id SERIAL PRIMARY KEY,
    owner VARCHAR(30) NOT NULL,
    ownerID INTEGER NOT NULL,
    title VARCHAR(255),
    start TIMESTAMP NOT NULL,
    finish TIMESTAMP NOT NULL,
    notify TIMESTAMP NOT NULL,
    CONSTRAINT fk_owner FOREIGN KEY (owner) REFERENCES users(name),
    CONSTRAINT fk_ownerID FOREIGN KEY(ownerID) REFERENCES users(id)
);

CREATE OR REPLACE FUNCTION NEW_EVENT (
    owner VARCHAR(30),
    start TIMESTAMP,
    finish TIMESTAMP,
    notify TIMESTAMP,
    title VARCHAR(255)

) RETURNS INTEGER AS $$
    DECLARE
        identifier INTEGER := 0;
        eventID INTEGER := 0;
    BEGIN
        SELECT users.id FROM users WHERE users.name = $1 INTO identifier;

        IF identifier ISNULL OR identifier = 0 THEN
            INSERT INTO users (name) VALUES($1) RETURNING users.id INTO identifier;
        END IF;

        INSERT INTO
            events (owner, ownerID, start, finish, notify, title)
        VALUES
            (owner, identifier, $2, $3, $4, $5)
        RETURNING events.id INTO eventID;

        RETURN eventID;
    END; $$

    LANGUAGE 'plpgsql';
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE events;
DROP TABLE users;
-- +goose StatementEnd
