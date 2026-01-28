CREATE TABLE collaborator (
    work_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (work_id, user_id),
    FOREIGN KEY (work_id) REFERENCES work(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
);
