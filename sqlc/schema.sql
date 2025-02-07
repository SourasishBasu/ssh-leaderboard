CREATE TABLE participants (
    id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
    game1 integer,
    game2 integer,
    game3 integer,
    name text NOT NULL,
    PRIMARY KEY (id)
);