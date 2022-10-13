-- +goose Up
-- +goose StatementBegin
CREATE TABLE owners
(
    ownerId SERIAL PRIMARY KEY
);

CREATE TABLE events
(
    ID SERIAL PRIMARY KEY,
    ownerID INTEGER,
    startTime TIMESTAMP,
    finishTime TIMESTAMP,
    title VARCHAR(255),
    CONSTRAINT fk_ownerId FOREIGN KEY(ownerId) REFERENCES owners(ownerId)
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE events;
DROP TABLE owners;
-- +goose StatementEnd
