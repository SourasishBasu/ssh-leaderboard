CREATE TABLE completed_questions (
    id integer NOT NULL,
    team_id text UNIQUE,
    question_status text UNIQUE,
    completed_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

CREATE TABLE team_points (
    team_id text NOT NULL,
    total_points integer DEFAULT 0,
    PRIMARY KEY (team_id)
);

CREATE TABLE teams (
    team_id text NOT NULL,
    team_name text UNIQUE NOT NULL,
    password_hash text NOT NULL,
    PRIMARY KEY (team_id)
);
