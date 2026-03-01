CREATE TABLE IF NOT EXISTS movies (
                                      id     SERIAL PRIMARY KEY,
                                      title  VARCHAR(255) NOT NULL,
    genre  VARCHAR(100) NOT NULL,
    budget BIGINT       NOT NULL DEFAULT 0
    );

CREATE TABLE IF NOT EXISTS technicians (
                                           id       SERIAL PRIMARY KEY,
                                           movie_id INT          NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    name     VARCHAR(255) NOT NULL,
    role     VARCHAR(100) NOT NULL
    );

-- Seed data
INSERT INTO movies (title, genre, budget) VALUES
                                              ('SAW', 'horror', 500000),
                                              ('The Matrix', 'sci-fi', 63000000),
                                              ('Inception', 'thriller', 160000000)
    ON CONFLICT DO NOTHING;

INSERT INTO technicians (movie_id, name, role) VALUES
                                                   (1, 'James Wan', 'Director'),
                                                   (1, 'Leigh Whannell', 'Writer'),
                                                   (2, 'Lana Wachowski', 'Director'),
                                                   (3, 'Christopher Nolan', 'Director')
    ON CONFLICT DO NOTHING;