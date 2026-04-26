CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL UNIQUE,
    age        INTEGER      NOT NULL DEFAULT 0,
    gender     VARCHAR(10)  NOT NULL DEFAULT 'unknown',
    birth_date DATE         NOT NULL DEFAULT '2000-01-01',
    created_at TIMESTAMP    NOT NULL DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS user_friends (
                                            user_id   INTEGER REFERENCES users(id) ON DELETE CASCADE,
    friend_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, friend_id),
    CONSTRAINT no_self_friend CHECK (user_id <> friend_id)
    );

INSERT INTO users (name, email, age, gender, birth_date) VALUES
     ('Alice Johnson',   'alice@example.com',   28, 'female', '1996-03-12'),
     ('Bob Smith',       'bob@example.com',     34, 'male',   '1990-07-24'),
     ('Carol White',     'carol@example.com',   22, 'female', '2002-01-05'),
     ('David Brown',     'david@example.com',   45, 'male',   '1979-11-30'),
     ('Eva Martinez',    'eva@example.com',     31, 'female', '1993-06-18'),
     ('Frank Wilson',    'frank@example.com',   27, 'male',   '1997-09-09'),
     ('Grace Lee',       'grace@example.com',   29, 'female', '1995-04-22'),
     ('Henry Taylor',    'henry@example.com',   38, 'male',   '1986-12-14'),
     ('Iris Anderson',   'iris@example.com',    24, 'female', '2000-08-03'),
     ('Jack Thomas',     'jack@example.com',    33, 'male',   '1991-02-27'),
     ('Karen Jackson',   'karen@example.com',   26, 'female', '1998-05-16'),
     ('Leo Harris',      'leo@example.com',     41, 'male',   '1983-10-07'),
     ('Mia Martin',      'mia@example.com',     23, 'female', '2001-03-29'),
     ('Noah Garcia',     'noah@example.com',    36, 'male',   '1988-07-11'),
     ('Olivia Davis',    'olivia@example.com',  30, 'female', '1994-01-19'),
     ('Peter Rodriguez', 'peter@example.com',   43, 'male',   '1981-09-02'),
     ('Quinn Wilson',    'quinn@example.com',   25, 'female', '1999-06-25'),
     ('Ryan Moore',      'ryan@example.com',    32, 'male',   '1992-11-08'),
     ('Sophia Taylor',   'sophia@example.com',  27, 'female', '1997-04-14'),
     ('Tom Anderson',    'tom@example.com',     39, 'male',   '1985-08-21')
    ON CONFLICT DO NOTHING;

-- user 1 (Alice) and user 2 (Bob) share 3 common friends: 3 (Carol), 4 (David), 5 (Eva)
-- user 3 (Carol) and user 4 (David) share 3 common friends: 1 (Alice), 2 (Bob), 5 (Eva)
INSERT INTO user_friends (user_id, friend_id) VALUES
      (1,2),(2,1),
      (1,3),(3,1),
      (1,4),(4,1),
      (1,5),(5,1),
      (2,3),(3,2),
      (2,4),(4,2),
      (2,5),(5,2),
      (3,4),(4,3),
      (3,5),(5,3),
      (4,5),(5,4),
      (4,6),(6,4),
      (5,6),(6,5),
      (6,7),(7,6),
      (7,8),(8,7),
      (8,9),(9,8),
      (9,10),(10,9),
      (10,11),(11,10),
      (11,12),(12,11),
      (12,13),(13,12),
      (13,14),(14,13),
      (14,15),(15,14),
      (15,16),(16,15)
    ON CONFLICT DO NOTHING;
