CREATE TABLE links (
    id SERIAL PRIMARY KEY, 
    user_id INTEGER NOT NULL,
    url TEXT,
    short_url TEXT,
    click_count INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL, 
    updated_at TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX links_url ON links (url);
CREATE UNIQUE INDEX links_short_url ON links (short_url);
CREATE UNIQUE INDEX links_user_id ON links (user_id);